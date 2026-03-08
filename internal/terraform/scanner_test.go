package terraform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseVariableBlocks(t *testing.T) {
	content := `
variable "project_name" {
  type        = string
  description = "Tenant project identifier"
}

variable "environment" {
  type        = string
  default     = "production"
  description = "Environment type"
}

variable "tier" {
  type        = string
  default     = "standard"
  description = "Tenant tier"
}

variable "vpc_cidr" {
  type        = string
  default     = "10.0.0.0/16"
  description = "VPC CIDR block"
}
`
	vars := parseVariableBlocks(content)
	if len(vars) != 4 {
		t.Fatalf("expected 4 variables, got %d", len(vars))
	}

	tests := []struct {
		idx         int
		name        string
		hasDefault  bool
		defaultVal  string
		description string
	}{
		{0, "project_name", false, "", "Tenant project identifier"},
		{1, "environment", true, "production", "Environment type"},
		{2, "tier", true, "standard", "Tenant tier"},
		{3, "vpc_cidr", true, "10.0.0.0/16", "VPC CIDR block"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := vars[tt.idx]
			if v.Name != tt.name {
				t.Errorf("name = %q, want %q", v.Name, tt.name)
			}
			if v.HasDefault != tt.hasDefault {
				t.Errorf("hasDefault = %v, want %v", v.HasDefault, tt.hasDefault)
			}
			if v.Default != tt.defaultVal {
				t.Errorf("default = %q, want %q", v.Default, tt.defaultVal)
			}
			if v.Description != tt.description {
				t.Errorf("description = %q, want %q", v.Description, tt.description)
			}
		})
	}
}

func TestParseModuleBlocks(t *testing.T) {
	content := `
module "networking" {
  source = "./modules/networking"
  vpc_cidr = var.vpc_cidr
}

module "database" {
  source = "./modules/database"
  instance_class = var.db_instance_class
}
`
	modules := parseModuleBlocks(content)
	if len(modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(modules))
	}
	if modules[0].Name != "networking" {
		t.Errorf("first module name = %q, want 'networking'", modules[0].Name)
	}
	if modules[0].Source != "./modules/networking" {
		t.Errorf("first module source = %q, want './modules/networking'", modules[0].Source)
	}
	if modules[1].Name != "database" {
		t.Errorf("second module name = %q, want 'database'", modules[1].Name)
	}
}

func TestScanVariables(t *testing.T) {
	dir := t.TempDir()
	content := `
variable "project_name" {
  type        = string
  description = "Tenant project identifier"
}

variable "environment" {
  type        = string
  default     = "production"
  description = "Environment type"
}
`
	if err := os.WriteFile(filepath.Join(dir, "variables.tf"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner(dir)
	vars, err := scanner.ScanVariables()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vars) != 2 {
		t.Fatalf("expected 2 variables, got %d", len(vars))
	}
}

func TestScanVariables_NoTfFiles(t *testing.T) {
	dir := t.TempDir()
	scanner := NewScanner(dir)
	_, err := scanner.ScanVariables()
	if err == nil {
		t.Fatal("expected error for directory with no .tf files")
	}
}

func TestScanModules(t *testing.T) {
	dir := t.TempDir()
	content := `
module "compute" {
  source = "./modules/compute"
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner(dir)
	modules, err := scanner.ScanModules()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(modules))
	}
	if modules[0].Name != "compute" {
		t.Errorf("module name = %q, want 'compute'", modules[0].Name)
	}
}

func TestDetectProjectMode(t *testing.T) {
	dir := t.TempDir()
	content := `
module "networking" {
  source = "./modules/networking"
}

variable "name" {
  type = string
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner(dir)
	mode, err := scanner.DetectProjectMode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mode != "root" {
		t.Errorf("mode = %q, want 'root'", mode)
	}
}
