package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GenerateTfvars writes a valid .tfvars file with the given variables.
func GenerateTfvars(variables map[string]string, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory for tfvars: %w", err)
	}

	var lines []string
	// Sort keys for deterministic output
	keys := make([]string, 0, len(variables))
	for key := range variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := variables[key]
		lines = append(lines, fmt.Sprintf("%s = %q", key, value))
	}

	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing tfvars file: %w", err)
	}
	return nil
}

// GenerateBackendOverride writes a backend override file that points
// terraform state to the tenant's isolated state directory.
func GenerateBackendOverride(stateDir string, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating directory for backend override: %w", err)
	}

	content := fmt.Sprintf(`terraform {
  backend "local" {
    path = %q
  }
}
`, filepath.Join(stateDir, "terraform.tfstate"))

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing backend override file: %w", err)
	}
	return nil
}
