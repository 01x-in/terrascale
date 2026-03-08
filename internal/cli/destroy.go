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

var (
	destroyAutoApprove bool
	destroyKeepState   bool
)

func init() {
	destroyCmd.Flags().BoolVar(&destroyAutoApprove, "auto-approve", false, "Skip confirmation prompt")
	destroyCmd.Flags().BoolVar(&destroyKeepState, "keep-state", false, "Preserve state directory after destruction")
	rootCmd.AddCommand(destroyCmd)
}

var destroyCmd = &cobra.Command{
	Use:   "destroy <slug>",
	Short: "Destroy a tenant's infrastructure",
	Long: `Destroy all infrastructure for a tenant, clean up state,
and mark the tenant as destroyed in the registry.`,
	Args: cobra.ExactArgs(1),
	RunE: runDestroy,
}

func runDestroy(cmd *cobra.Command, args []string) error {
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
		return fmt.Errorf("tenant %q is already destroyed", slug)
	}

	// Confirmation
	if !destroyAutoApprove {
		fmt.Printf("\n  This will destroy ALL infrastructure for tenant %q.\n", slug)
		fmt.Printf("  This action cannot be undone.\n\n")
		confirmed, err := ui.Confirm(
			fmt.Sprintf("Type '%s' to confirm destruction", slug),
			slug,
		)
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Warn("Destruction cancelled. Slug did not match.")
			return nil
		}
	}

	// Find terraform binary
	tfBinary, err := terraform.FindTerraformBinary()
	if err != nil {
		return err
	}

	// Update status to destroying
	if err := registry.UpdateTenantStatus(cfg, slug, config.StatusDestroying); err != nil {
		return err
	}
	config.SaveConfig(cfg, configPath)

	// Determine working directory and state
	tfDir := cwd
	if cfg.Project.TerraformDir != "" && cfg.Project.TerraformDir != "." {
		tfDir = filepath.Join(cwd, cfg.Project.TerraformDir)
	}

	stateDir := filepath.Join(cwd, tenant.StatePath)
	absStateDir, err := filepath.Abs(stateDir)
	if err != nil {
		return fmt.Errorf("resolving state path: %w", err)
	}

	// Generate backend override for this tenant's state
	overridePath := filepath.Join(tfDir, "backend_override.tf")
	if err := terraform.GenerateBackendOverride(absStateDir, overridePath); err != nil {
		return fmt.Errorf("generating backend override: %w", err)
	}
	defer os.Remove(overridePath)

	// Generate tfvars for destroy
	tfvarsPath := filepath.Join(stateDir, "tenant.tfvars")
	if _, err := os.Stat(tfvarsPath); os.IsNotExist(err) {
		// Regenerate tfvars if missing
		if err := terraform.GenerateTfvars(tenant.Variables, tfvarsPath); err != nil {
			return fmt.Errorf("generating tfvars: %w", err)
		}
	}

	executor := terraform.NewExecutor(tfDir, tfBinary)

	// Run terraform init
	ui.Step("Running terraform init...")
	if err := executor.Init(nil); err != nil {
		registry.UpdateTenantStatus(cfg, slug, config.StatusFailed)
		config.SaveConfig(cfg, configPath)
		return fmt.Errorf("terraform init failed during destroy for tenant %q: %w", slug, err)
	}

	// Run terraform destroy
	ui.Step("Running terraform destroy...")
	if err := executor.Destroy(tfvarsPath, true); err != nil {
		registry.UpdateTenantStatus(cfg, slug, config.StatusFailed)
		config.SaveConfig(cfg, configPath)
		return fmt.Errorf("terraform destroy failed for tenant %q. State has been preserved at %s for debugging.\n%w", slug, tenant.StatePath, err)
	}
	ui.Success("Terraform destroy complete")

	// Update status
	if err := registry.UpdateTenantStatus(cfg, slug, config.StatusDestroyed); err != nil {
		return err
	}

	// Clean up state directory
	if !destroyKeepState {
		if err := terraform.RemoveStateDir(cwd, slug); err != nil {
			ui.Warn(fmt.Sprintf("Could not remove state directory: %v", err))
		} else {
			ui.Success("State directory cleaned up")
		}
	} else {
		ui.Info("State directory preserved (--keep-state)")
	}

	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Println()
	ui.Success(fmt.Sprintf("Tenant %q destroyed successfully.", slug))
	return nil
}
