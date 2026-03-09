package cli

import (
	"encoding/json"
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
	inspectOutputsOnly bool
	inspectRefresh     bool
	inspectJSON        bool
)

func init() {
	inspectCmd.Flags().BoolVar(&inspectOutputsOnly, "outputs-only", false, "Show only tenant outputs")
	inspectCmd.Flags().BoolVar(&inspectRefresh, "refresh", false, "Run terraform refresh before showing")
	inspectCmd.Flags().BoolVar(&inspectJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(inspectCmd)
}

var inspectCmd = &cobra.Command{
	Use:   "inspect <slug>",
	Short: "Show detailed information about a tenant",
	Long:  `Display all variables, outputs, state path, and timestamps for a tenant.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runInspect,
}

func runInspect(cmd *cobra.Command, args []string) error {
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

	// Optionally refresh state
	if inspectRefresh && tenant.Status == config.StatusActive {
		tfBinary, err := terraform.FindTerraformBinary()
		if err != nil {
			return err
		}

		tfDir := cwd
		if cfg.Project.TerraformDir != "" && cfg.Project.TerraformDir != "." {
			tfDir = filepath.Join(cwd, cfg.Project.TerraformDir)
		}

		executor := terraform.NewExecutor(tfDir, tfBinary)
		ui.Step("Refreshing tenant state...")
		if err := executor.Refresh(); err != nil {
			ui.Warn(fmt.Sprintf("Refresh failed: %v", err))
		} else {
			// Re-capture outputs after refresh
			outputs, err := executor.Output()
			if err == nil {
				if err := registry.UpdateTenantOutputs(cfg, slug, outputs); err != nil {
					ui.Warn(fmt.Sprintf("Could not update outputs: %v", err))
				} else {
					config.SaveConfig(cfg, configPath)
					tenant, _ = registry.GetTenant(cfg, slug)
				}
			}
			ui.Success("State refreshed")
		}
	}

	if inspectJSON {
		return printTenantJSON(tenant)
	}

	if inspectOutputsOnly {
		return printOutputsOnly(tenant)
	}

	return printTenantDetail(tenant)
}

func printTenantJSON(tenant *config.Tenant) error {
	data, err := json.MarshalIndent(tenant, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func printOutputsOnly(tenant *config.Tenant) error {
	if len(tenant.Outputs) == 0 {
		fmt.Println("No outputs captured for this tenant.")
		return nil
	}
	for key, value := range tenant.Outputs {
		fmt.Printf("%s = %s\n", key, value)
	}
	return nil
}

func printTenantDetail(tenant *config.Tenant) error {
	fmt.Println()
	fmt.Printf("  %s\n", ui.Bold(tenant.Slug))
	fmt.Println()
	fmt.Printf("  Name:        %s\n", tenant.Name)
	fmt.Printf("  Environment: %s\n", tenant.Environment)
	fmt.Printf("  Status:      %s\n", tenant.Status)
	fmt.Printf("  State Path:  %s\n", tenant.StatePath)
	fmt.Printf("  Created:     %s\n", tenant.CreatedAt.Format("2006-01-02 15:04:05 UTC"))
	fmt.Printf("  Updated:     %s\n", tenant.UpdatedAt.Format("2006-01-02 15:04:05 UTC"))

	if len(tenant.Variables) > 0 {
		fmt.Println()
		fmt.Println("  Variables:")
		for key, value := range tenant.Variables {
			fmt.Printf("    %s = %s\n", key, value)
		}
	}

	if len(tenant.Outputs) > 0 {
		fmt.Println()
		fmt.Println("  Outputs:")
		for key, value := range tenant.Outputs {
			fmt.Printf("    %s = %s\n", key, value)
		}
	}

	fmt.Println()
	return nil
}
