package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "terrascale",
	Short: "Multi-tenant lifecycle manager for Terraform",
	Long: `TerraScale wraps your existing Terraform project to provision, manage,
and destroy isolated tenants — each with their own state file and configuration.

No code duplication. No manual state management. No custom scripts.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
