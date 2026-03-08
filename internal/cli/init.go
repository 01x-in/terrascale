package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/01x-in/terrascale/internal/config"
	"github.com/01x-in/terrascale/internal/terraform"
	"github.com/01x-in/terrascale/internal/ui"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize TerraScale in a Terraform project",
	Long: `Scan the current Terraform project, discover variables and modules,
and generate a terrascale.yaml configuration file.`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	// Check if already initialized
	configPath := filepath.Join(cwd, config.ConfigFileName)
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("terrascale.yaml already exists in this directory. Delete it first if you want to reinitialize")
	}

	// Scan the terraform project
	scanner := terraform.NewScanner(cwd)
	variables, err := scanner.ScanVariables()
	if err != nil {
		return fmt.Errorf("scanning project: %w", err)
	}

	ui.Success(fmt.Sprintf("Found %d variables in Terraform project", len(variables)))

	modules, err := scanner.ScanModules()
	if err != nil {
		// Not fatal — project might not have modules
		modules = nil
	}
	if len(modules) > 0 {
		ui.Info(fmt.Sprintf("Found %d module(s): %s", len(modules), moduleNames(modules)))
	}

	// Detect project mode
	mode, _ := scanner.DetectProjectMode()

	// Interactive prompts
	projectName, err := ui.InputString("Project name", filepath.Base(cwd))
	if err != nil {
		return err
	}
	if projectName == "" {
		projectName = filepath.Base(cwd)
	}

	// Ask which variables are tenant-specific
	varNames := make([]string, len(variables))
	for i, v := range variables {
		label := v.Name
		if v.Description != "" {
			label = fmt.Sprintf("%s (%s)", v.Name, v.Description)
		}
		varNames[i] = label
	}

	tenantVarNames, err := ui.MultiSelect(
		"Select variables that CHANGE per tenant (others will be shared)",
		varNames,
	)
	if err != nil {
		return err
	}

	// Build config
	cfg := config.DefaultConfig(projectName)
	cfg.Project.Mode = mode

	// Separate tenant vs shared variables
	selectedSet := make(map[string]bool)
	for _, selected := range tenantVarNames {
		// Extract name from "name (description)" format
		name := strings.Split(selected, " (")[0]
		selectedSet[name] = true
	}

	for _, v := range variables {
		if selectedSet[v.Name] {
			varDef := config.VariableDef{
				Name:     v.Name,
				Type:     cleanType(v.Type),
				Required: !v.HasDefault,
			}
			if v.Default != "" {
				varDef.Default = v.Default
			}
			if v.Description != "" {
				varDef.Prompt = v.Description
			}
			cfg.TenantSpec.TenantVariables = append(cfg.TenantSpec.TenantVariables, varDef)
		} else if v.HasDefault {
			cfg.TenantSpec.SharedVariables[v.Name] = v.Default
		}
	}

	// Set up default tiers
	cfg.Tiers = map[string]config.TierPreset{
		"basic": {
			VpcMode:         "shared",
			DbInstanceClass: "db.t3.micro",
		},
		"standard": {
			VpcMode:         "shared",
			DbInstanceClass: "db.t3.small",
		},
		"premium": {
			VpcMode:         "dedicated",
			DbInstanceClass: "db.t3.medium",
		},
	}

	// Capture outputs
	cfg.TenantSpec.Outputs = []string{}

	// Save config
	if err := config.SaveConfig(cfg, configPath); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	ui.Success(fmt.Sprintf("Generated %s", config.ConfigFileName))

	// Create .terrascale directory
	if err := terraform.EnsureTerrascaleDir(cwd); err != nil {
		return err
	}
	ui.Success("Created .terrascale/ directory")

	// Update .gitignore
	if err := updateGitignore(cwd); err != nil {
		ui.Warn(fmt.Sprintf("Could not update .gitignore: %v", err))
	} else {
		ui.Success("Updated .gitignore")
	}

	fmt.Println()
	ui.Success("TerraScale initialized! Next steps:")
	fmt.Println("  1. Review terrascale.yaml and adjust variable settings")
	fmt.Println("  2. Run 'terrascale add <slug>' to provision your first tenant")

	return nil
}

func moduleNames(modules []terraform.DiscoveredModule) string {
	names := make([]string, len(modules))
	for i, m := range modules {
		names[i] = m.Name
	}
	return strings.Join(names, ", ")
}

func cleanType(t string) string {
	if t == "" {
		return "string"
	}
	return strings.TrimSpace(t)
}

func updateGitignore(dir string) error {
	gitignorePath := filepath.Join(dir, ".gitignore")
	entry := ".terrascale/state/"

	content, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Check if already present
	if strings.Contains(string(content), entry) {
		return nil
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Add newline if file doesn't end with one
	if len(content) > 0 && content[len(content)-1] != '\n' {
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}

	_, err = f.WriteString(entry + "\n")
	return err
}
