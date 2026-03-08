package terraform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateTfvars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tenant.tfvars")

	vars := map[string]string{
		"project_name": "my-project",
		"environment":  "production",
		"vpc_cidr":     "10.0.0.0/16",
	}

	if err := GenerateTfvars(vars, path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}

	str := string(content)
	// Check variables are present and sorted
	if !strings.Contains(str, `environment = "production"`) {
		t.Error("missing environment variable")
	}
	if !strings.Contains(str, `project_name = "my-project"`) {
		t.Error("missing project_name variable")
	}
	if !strings.Contains(str, `vpc_cidr = "10.0.0.0/16"`) {
		t.Error("missing vpc_cidr variable")
	}

	// Verify sorted order
	envIdx := strings.Index(str, "environment")
	projIdx := strings.Index(str, "project_name")
	vpcIdx := strings.Index(str, "vpc_cidr")
	if envIdx > projIdx || projIdx > vpcIdx {
		t.Error("variables should be sorted alphabetically")
	}
}

func TestGenerateTfvars_EmptyVars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.tfvars")

	if err := GenerateTfvars(map[string]string{}, path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}
	if strings.TrimSpace(string(content)) != "" {
		t.Errorf("expected empty file, got %q", string(content))
	}
}

func TestGenerateTfvars_CreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "dir", "tenant.tfvars")

	if err := GenerateTfvars(map[string]string{"a": "b"}, path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file was not created")
	}
}

func TestGenerateBackendOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "backend_override.tf")
	stateDir := "/some/state/dir"

	if err := GenerateBackendOverride(stateDir, path); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading generated file: %v", err)
	}

	str := string(content)
	if !strings.Contains(str, `backend "local"`) {
		t.Error("missing backend local block")
	}
	if !strings.Contains(str, "/some/state/dir/terraform.tfstate") {
		t.Error("missing state path")
	}
}
