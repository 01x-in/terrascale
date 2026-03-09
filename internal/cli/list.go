package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/01x-in/terrascale/internal/config"
	"github.com/01x-in/terrascale/internal/registry"
	"github.com/01x-in/terrascale/internal/ui"
)

var (
	listStatus      string
	listEnvironment string
	listJSON        bool
)

func init() {
	listCmd.Flags().StringVar(&listStatus, "status", "", "Filter by status (active, destroyed, failed, provisioning)")
	listCmd.Flags().StringVar(&listEnvironment, "environment", "", "Filter by environment (development, uat, staging, demo, production)")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output in JSON format")
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tenants",
	Long:  `Display a table of all tenants with their status and environment.`,
	RunE:  runList,
}

func runList(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	configPath := filepath.Join(cwd, config.ConfigFileName)
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("terrascale.yaml not found. Run 'terrascale init' first to set up your project")
	}

	filters := registry.TenantFilters{
		Status:      listStatus,
		Environment: listEnvironment,
	}
	tenants := registry.ListTenants(cfg, filters)

	if listJSON {
		return printTenantsJSON(tenants)
	}

	if len(tenants) == 0 {
		fmt.Println("No tenants found. Run 'terrascale add <slug>' to provision your first tenant.")
		return nil
	}

	headers := []string{"SLUG", "NAME", "ENVIRONMENT", "STATUS", "CREATED"}
	var rows [][]string
	for _, t := range tenants {
		created := t.CreatedAt.Format("2006-01-02")
		rows = append(rows, []string{
			t.Slug,
			t.Name,
			t.Environment,
			string(t.Status),
			created,
		})
	}

	ui.PrintTable(os.Stdout, headers, rows)
	return nil
}

func printTenantsJSON(tenants []config.Tenant) error {
	data, err := json.MarshalIndent(tenants, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}
