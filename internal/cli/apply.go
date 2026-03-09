package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/01x-in/terrascale/internal/config"
	"github.com/01x-in/terrascale/internal/registry"
	"github.com/01x-in/terrascale/internal/terraform"
	"github.com/01x-in/terrascale/internal/ui"
)

var applyAutoApprove bool

func init() {
	applyCmd.Flags().BoolVar(&applyAutoApprove, "auto-approve", false, "Skip confirmation prompts")
	rootCmd.AddCommand(applyCmd)
}

var applyCmd = &cobra.Command{
	Use:   "apply <slug>",
	Short: "Re-apply infrastructure for an existing tenant",
	Long: `Re-run terraform plan and apply for an existing tenant.
Useful for retrying a failed provisioning or picking up configuration changes.`,
	Args: cobra.ExactArgs(1),
	RunE: runApply,
}

func runApply(cmd *cobra.Command, args []string) error {
	slug := args[0]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	configPath := filepath.Join(cwd, config.ConfigFileName)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("terrascale.yaml not found. Run 'terrascale init' first to set up your project")
	}

	tenant, err := registry.GetTenant(cfg, slug)
	if err != nil {
		return err
	}

	if tenant.Status == config.StatusDestroyed {
		return fmt.Errorf("tenant %q is destroyed. Use 'terrascale add %s' to provision a new one", slug, slug)
	}

	// Find terraform binary
	tfBinary, err := terraform.FindTerraformBinary()
	if err != nil {
		return err
	}

	// Show AWS identity and confirm
	if !applyAutoApprove {
		identity, err := terraform.GetAWSIdentity()
		if err != nil {
			return err
		}
		fmt.Printf("\n  AWS Profile:  %s\n", identity.Profile)
		fmt.Printf("  Account ID:   %s\n", identity.AccountID)
		fmt.Printf("  Identity:     %s\n\n", identity.ARN)

		confirmed, err := ui.ConfirmYesNo(fmt.Sprintf("Re-apply tenant %q on this AWS account?", slug))
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Warn("Aborted.")
			return nil
		}
	}

	// Determine working directory
	tfDir := cwd
	if cfg.Project.TerraformDir != "" && cfg.Project.TerraformDir != "." {
		tfDir = filepath.Join(cwd, cfg.Project.TerraformDir)
	}

	stateDir := filepath.Join(cwd, tenant.StatePath)
	absStateDir, err := filepath.Abs(stateDir)
	if err != nil {
		return fmt.Errorf("resolving state path: %w", err)
	}

	// Regenerate backend override
	overridePath := filepath.Join(tfDir, "backend_override.tf")
	if err := terraform.GenerateBackendOverride(absStateDir, overridePath); err != nil {
		return fmt.Errorf("generating backend override: %w", err)
	}
	defer os.Remove(overridePath)

	// Regenerate tfvars from saved variables
	tfvarsPath := filepath.Join(stateDir, "tenant.tfvars")
	if err := terraform.GenerateTfvars(tenant.Variables, tfvarsPath); err != nil {
		return fmt.Errorf("generating tfvars: %w", err)
	}
	ui.Success("Generated tenant.tfvars")

	executor := terraform.NewExecutor(tfDir, tfBinary)

	// Mark as updating
	if err := registry.UpdateTenantStatus(cfg, slug, config.StatusUpdating); err != nil {
		return err
	}
	config.SaveConfig(cfg, configPath)

	// terraform init
	ui.Step("Running terraform init...")
	if err := executor.Init(nil); err != nil {
		registry.UpdateTenantStatus(cfg, slug, config.StatusFailed)
		config.SaveConfig(cfg, configPath)
		return fmt.Errorf("terraform init failed for tenant %q: %w", slug, err)
	}
	ui.Success("Terraform initialized")

	// terraform plan
	ui.Step("Running terraform plan...")
	planResult, err := executor.Plan(tfvarsPath)
	if err != nil {
		registry.UpdateTenantStatus(cfg, slug, config.StatusFailed)
		config.SaveConfig(cfg, configPath)
		return fmt.Errorf("terraform plan failed for tenant %q: %w", slug, err)
	}
	fmt.Printf("\n  Plan: %d to add, %d to change, %d to destroy\n\n",
		planResult.ToAdd, planResult.ToChange, planResult.ToDestroy)

	// Confirm apply
	if !applyAutoApprove {
		confirmed, err := ui.ConfirmYesNo("Apply this plan?")
		if err != nil {
			return err
		}
		if !confirmed {
			registry.UpdateTenantStatus(cfg, slug, config.StatusFailed)
			config.SaveConfig(cfg, configPath)
			ui.Warn("Aborted.")
			return nil
		}
	}

	// terraform apply
	ui.Step("Running terraform apply...")
	applyResult, err := executor.Apply(tfvarsPath, true)
	if err != nil {
		registry.UpdateTenantStatus(cfg, slug, config.StatusFailed)
		config.SaveConfig(cfg, configPath)
		return fmt.Errorf("terraform apply failed for tenant %q. State has been preserved at %s for debugging.\n%w", slug, tenant.StatePath, err)
	}
	if !applyResult.Success {
		registry.UpdateTenantStatus(cfg, slug, config.StatusFailed)
		config.SaveConfig(cfg, configPath)
		return fmt.Errorf("terraform apply failed for tenant %q", slug)
	}
	ui.Success("Terraform apply complete")

	// Capture outputs
	outputs, err := executor.Output()
	if err != nil {
		ui.Warn(fmt.Sprintf("Could not capture outputs: %v", err))
		outputs = map[string]string{}
	}

	// Update registry
	if err := registry.UpdateTenantStatus(cfg, slug, config.StatusActive); err != nil {
		return err
	}
	if err := registry.UpdateTenantOutputs(cfg, slug, outputs); err != nil {
		return err
	}
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Tenant %q applied successfully!", slug))

	if len(outputs) > 0 {
		fmt.Println()
		for key, value := range outputs {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}

	fmt.Println()
	fmt.Printf("  Run 'terrascale inspect %s' to see full details.\n", slug)
	return nil
}
