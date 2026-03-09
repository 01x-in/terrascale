package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/01x-in/terrascale/internal/config"
	"github.com/01x-in/terrascale/internal/registry"
	"github.com/01x-in/terrascale/internal/terraform"
	"github.com/01x-in/terrascale/internal/ui"
)

var (
	addVars        []string
	addName        string
	addEnv         string
	addAutoApprove bool
)

func init() {
	addCmd.Flags().StringArrayVar(&addVars, "var", nil, "Variable in key=value format (can be repeated)")
	addCmd.Flags().StringVar(&addName, "name", "", "Display name for the tenant")
	addCmd.Flags().StringVar(&addEnv, "environment", "production", "Environment type")
	addCmd.Flags().BoolVar(&addAutoApprove, "auto-approve", false, "Skip confirmation prompt")
	rootCmd.AddCommand(addCmd)
}

var addCmd = &cobra.Command{
	Use:   "add <slug>",
	Short: "Provision a new tenant",
	Long: `Create a new tenant with isolated state and configuration.
The slug must be lowercase alphanumeric with hyphens.`,
	Args: cobra.ExactArgs(1),
	RunE: runAdd,
}

func runAdd(cmd *cobra.Command, args []string) error {
	slug := args[0]

	// Validate slug
	if err := config.ValidateSlug(slug); err != nil {
		return fmt.Errorf("invalid slug: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	// Load config
	configPath := filepath.Join(cwd, config.ConfigFileName)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("terrascale.yaml not found. Run 'terrascale init' first to set up your project")
	}

	// Check for duplicate
	if registry.TenantExists(cfg, slug) {
		return fmt.Errorf("tenant %q already exists. Use a different slug or destroy the existing tenant first", slug)
	}

	// Find terraform binary
	tfBinary, err := terraform.FindTerraformBinary()
	if err != nil {
		return err
	}

	// Show AWS identity and confirm before proceeding
	if !addAutoApprove {
		identity, err := terraform.GetAWSIdentity()
		if err != nil {
			return err
		}
		fmt.Printf("\n  AWS Profile:  %s\n", identity.Profile)
		fmt.Printf("  Account ID:   %s\n", identity.AccountID)
		fmt.Printf("  Identity:     %s\n\n", identity.ARN)

		confirmed, err := ui.ConfirmYesNo(fmt.Sprintf("Deploy tenant %q to this AWS account?", slug))
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Warn("Aborted.")
			return nil
		}
	}

	// Determine terraform working directory
	tfDir := cwd
	if cfg.Project.TerraformDir != "" && cfg.Project.TerraformDir != "." {
		tfDir = filepath.Join(cwd, cfg.Project.TerraformDir)
	}

	// Parse --var flags into map
	varMap := parseVarFlags(addVars)

	// Read shared variable values from terraform.tfvars (if it exists)
	allVars := make(map[string]string)
	if len(cfg.TenantSpec.SharedVariables) > 0 {
		tfvarsValues, err := terraform.ReadTfvars(filepath.Join(tfDir, "terraform.tfvars"))
		if err != nil {
			return fmt.Errorf("reading terraform.tfvars: %w", err)
		}
		for _, name := range cfg.TenantSpec.SharedVariables {
			if val, ok := tfvarsValues[name]; ok {
				allVars[name] = val
			}
			// If not in tfvars, skip — Terraform will use the variable's default
		}
	}

	// Apply provided variables (override defaults)
	for key, value := range varMap {
		allVars[key] = value
	}

	// Prompt for missing required variables
	for _, varDef := range cfg.TenantSpec.TenantVariables {
		if _, exists := allVars[varDef.Name]; exists {
			continue
		}
		if varDef.Required {
			prompt := varDef.Prompt
			if prompt == "" {
				prompt = fmt.Sprintf("Enter value for %s", varDef.Name)
			}
			if len(varDef.Options) > 0 {
				value, err := ui.SelectString(prompt, varDef.Options)
				if err != nil {
					return err
				}
				allVars[varDef.Name] = value
			} else {
				value, err := ui.InputString(prompt, varDef.Default)
				if err != nil {
					return err
				}
				if value == "" && varDef.Default != "" {
					value = varDef.Default
				}
				allVars[varDef.Name] = value
			}
		} else if varDef.Default != "" {
			allVars[varDef.Name] = varDef.Default
		}
	}

	// Create state directory
	stateDir, err := terraform.CreateStateDir(cwd, slug)
	if err != nil {
		return fmt.Errorf("creating state directory: %w", err)
	}
	statePath := terraform.GetStatePath(slug)
	ui.Success(fmt.Sprintf("Created state directory: %s", statePath))

	// Generate tenant.tfvars
	tfvarsPath := filepath.Join(stateDir, "tenant.tfvars")
	if err := terraform.GenerateTfvars(allVars, tfvarsPath); err != nil {
		return fmt.Errorf("generating tfvars: %w", err)
	}
	ui.Success("Generated tenant.tfvars")

	// Generate backend override
	absStateDir, err := filepath.Abs(stateDir)
	if err != nil {
		return fmt.Errorf("resolving state path: %w", err)
	}
	overridePath := filepath.Join(tfDir, "backend_override.tf")
	if err := terraform.GenerateBackendOverride(absStateDir, overridePath); err != nil {
		return fmt.Errorf("generating backend override: %w", err)
	}
	ui.Success("Generated backend override")

	// Create tenant record early (status: provisioning)
	now := time.Now().UTC()
	tenant := config.Tenant{
		Slug:        slug,
		Name:        addName,
		Environment: addEnv,
		Status:      config.StatusProvisioning,
		CreatedAt:   now,
		UpdatedAt:   now,
		Variables:   allVars,
		StatePath:   statePath + "/",
		AccountMode: "same",
	}
	if tenant.Name == "" {
		tenant.Name = slug
	}

	if err := registry.AddTenant(cfg, tenant); err != nil {
		return err
	}
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	// Run terraform init
	ui.Step("Running terraform init...")
	executor := terraform.NewExecutor(tfDir, tfBinary)
	if err := executor.Init(nil); err != nil {
		markFailed(cfg, configPath, slug)
		return fmt.Errorf("terraform init failed for tenant %q. State has been preserved at %s for debugging. Run 'terrascale inspect %s' to see current status.\n%w", slug, statePath, slug, err)
	}
	ui.Success("Terraform initialized")

	// Run terraform plan
	ui.Step("Running terraform plan...")
	planResult, err := executor.Plan(tfvarsPath)
	if err != nil {
		markFailed(cfg, configPath, slug)
		return fmt.Errorf("terraform plan failed for tenant %q. State has been preserved at %s for debugging.\n%w", slug, statePath, err)
	}

	fmt.Printf("\n  Plan: %d to add, %d to change, %d to destroy\n\n",
		planResult.ToAdd, planResult.ToChange, planResult.ToDestroy)

	// Confirm
	if !addAutoApprove {
		confirmed, err := ui.ConfirmYesNo("Apply this plan?")
		if err != nil {
			return err
		}
		if !confirmed {
			ui.Warn("Aborted. Tenant left in 'provisioning' state. Run 'terrascale destroy " + slug + "' to clean up.")
			return nil
		}
	}

	// Run terraform apply
	ui.Step("Running terraform apply...")
	applyResult, err := executor.Apply(tfvarsPath, true)
	if err != nil {
		markFailed(cfg, configPath, slug)
		return fmt.Errorf("terraform apply failed for tenant %q. State has been preserved at %s for debugging. Run 'terrascale inspect %s' to see current status.\n%w", slug, statePath, slug, err)
	}

	if !applyResult.Success {
		markFailed(cfg, configPath, slug)
		return fmt.Errorf("terraform apply failed for tenant %q. Check the output above for details", slug)
	}
	ui.Success("Terraform apply complete")

	// Capture outputs
	outputs, err := executor.Output()
	if err != nil {
		ui.Warn(fmt.Sprintf("Could not capture outputs: %v", err))
		outputs = map[string]string{}
	}

	// Save outputs to state directory (not in terrascale.yaml)
	if err := terraform.SaveOutputs(stateDir, outputs); err != nil {
		ui.Warn(fmt.Sprintf("Could not save outputs: %v", err))
	}

	// Update tenant status to active
	if err := registry.UpdateTenantStatus(cfg, slug, config.StatusActive); err != nil {
		return err
	}
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	// Clean up backend override
	os.Remove(overridePath)

	fmt.Println()
	ui.Success(fmt.Sprintf("Tenant %q provisioned successfully!", slug))

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

func parseVarFlags(vars []string) map[string]string {
	result := make(map[string]string)
	for _, v := range vars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

func markFailed(cfg *config.Config, configPath, slug string) {
	registry.UpdateTenantStatus(cfg, slug, config.StatusFailed)
	config.SaveConfig(cfg, configPath)
}
