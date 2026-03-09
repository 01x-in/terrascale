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

// ReadTfvars parses a .tfvars file and returns a map of variable name → value.
// Only simple string assignments (key = "value") are parsed; complex types are skipped.
func ReadTfvars(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("reading tfvars file: %w", err)
	}

	result := make(map[string]string)
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"`)
		result[key] = val
	}
	return result, nil
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
