package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_ValidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "terrascale.yaml")

	content := `version: "1"
project:
  name: test-project
  terraform_dir: "."
  mode: root
state:
  backend: local
tenant_spec:
  tenant_variables: []
tenants: []
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.Version != "1" {
		t.Errorf("expected version '1', got %q", cfg.Version)
	}
	if cfg.Project.Name != "test-project" {
		t.Errorf("expected project name 'test-project', got %q", cfg.Project.Name)
	}
	if cfg.Project.TerraformDir != "." {
		t.Errorf("expected terraform_dir '.', got %q", cfg.Project.TerraformDir)
	}
	if cfg.Project.Mode != "root" {
		t.Errorf("expected mode 'root', got %q", cfg.Project.Mode)
	}
	if cfg.State.Backend != "local" {
		t.Errorf("expected backend 'local', got %q", cfg.State.Backend)
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "terrascale.yaml")

	content := `{{{not valid yaml`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path/terrascale.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadConfig_ModuleMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "terrascale.yaml")

	content := `version: "1"
project:
  name: module-project
  terraform_dir: "."
  mode: module
  module: "./modules/tenant"
state:
  backend: local
tenants: []
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}
	if cfg.Project.Mode != "module" {
		t.Errorf("expected mode 'module', got %q", cfg.Project.Mode)
	}
	if cfg.Project.Module != "./modules/tenant" {
		t.Errorf("expected module './modules/tenant', got %q", cfg.Project.Module)
	}
}

func TestSaveConfig_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "terrascale.yaml")

	original := DefaultConfig("round-trip-test")
	original.TenantSpec.TenantVariables = []VariableDef{
		{Name: "project_name", Type: "string", Required: true},
	}
	if err := SaveConfig(original, path); err != nil {
		t.Fatalf("SaveConfig returned error: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if loaded.Version != original.Version {
		t.Errorf("version mismatch: got %q, want %q", loaded.Version, original.Version)
	}
	if loaded.Project.Name != original.Project.Name {
		t.Errorf("project name mismatch: got %q, want %q", loaded.Project.Name, original.Project.Name)
	}
	if loaded.Project.TerraformDir != original.Project.TerraformDir {
		t.Errorf("terraform_dir mismatch: got %q, want %q", loaded.Project.TerraformDir, original.Project.TerraformDir)
	}
	if loaded.State.Backend != original.State.Backend {
		t.Errorf("backend mismatch: got %q, want %q", loaded.State.Backend, original.State.Backend)
	}
	if len(loaded.TenantSpec.TenantVariables) != 1 {
		t.Fatalf("expected 1 tenant variable, got %d", len(loaded.TenantSpec.TenantVariables))
	}
	if loaded.TenantSpec.TenantVariables[0].Name != "project_name" {
		t.Errorf("expected tenant variable name 'project_name', got %q", loaded.TenantSpec.TenantVariables[0].Name)
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid root config",
			cfg: Config{
				Version: "1",
				Project: Project{Name: "test", TerraformDir: ".", Mode: "root"},
				State:   StateConfig{Backend: "local"},
			},
			wantErr: false,
		},
		{
			name: "valid module config",
			cfg: Config{
				Version: "1",
				Project: Project{Name: "test", TerraformDir: ".", Mode: "module", Module: "./mod"},
				State:   StateConfig{Backend: "local"},
			},
			wantErr: false,
		},
		{
			name: "valid s3 backend",
			cfg: Config{
				Version: "1",
				Project: Project{Name: "test", TerraformDir: ".", Mode: "root"},
				State:   StateConfig{Backend: "s3"},
			},
			wantErr: false,
		},
		{
			name: "missing version",
			cfg: Config{
				Project: Project{Name: "test", TerraformDir: ".", Mode: "root"},
				State:   StateConfig{Backend: "local"},
			},
			wantErr: true,
		},
		{
			name: "missing project name",
			cfg: Config{
				Version: "1",
				Project: Project{TerraformDir: ".", Mode: "root"},
				State:   StateConfig{Backend: "local"},
			},
			wantErr: true,
		},
		{
			name: "missing terraform_dir",
			cfg: Config{
				Version: "1",
				Project: Project{Name: "test", Mode: "root"},
				State:   StateConfig{Backend: "local"},
			},
			wantErr: true,
		},
		{
			name: "invalid mode",
			cfg: Config{
				Version: "1",
				Project: Project{Name: "test", TerraformDir: ".", Mode: "invalid"},
				State:   StateConfig{Backend: "local"},
			},
			wantErr: true,
		},
		{
			name: "module mode without module path",
			cfg: Config{
				Version: "1",
				Project: Project{Name: "test", TerraformDir: ".", Mode: "module"},
				State:   StateConfig{Backend: "local"},
			},
			wantErr: true,
		},
		{
			name: "missing backend",
			cfg: Config{
				Version: "1",
				Project: Project{Name: "test", TerraformDir: ".", Mode: "root"},
			},
			wantErr: true,
		},
		{
			name: "invalid backend",
			cfg: Config{
				Version: "1",
				Project: Project{Name: "test", TerraformDir: ".", Mode: "root"},
				State:   StateConfig{Backend: "gcs"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(&tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig("my-project")

	if cfg.Version != "1" {
		t.Errorf("expected version '1', got %q", cfg.Version)
	}
	if cfg.Project.Name != "my-project" {
		t.Errorf("expected project name 'my-project', got %q", cfg.Project.Name)
	}
	if cfg.Project.TerraformDir != "." {
		t.Errorf("expected terraform_dir '.', got %q", cfg.Project.TerraformDir)
	}
	if cfg.Project.Mode != "root" {
		t.Errorf("expected mode 'root', got %q", cfg.Project.Mode)
	}
	if cfg.State.Backend != "local" {
		t.Errorf("expected backend 'local', got %q", cfg.State.Backend)
	}
	if cfg.Tenants == nil {
		t.Error("expected Tenants to be initialized, got nil")
	}
	if len(cfg.Tenants) != 0 {
		t.Errorf("expected 0 tenants, got %d", len(cfg.Tenants))
	}

	// Verify it passes validation
	if err := ValidateConfig(cfg); err != nil {
		t.Errorf("DefaultConfig should pass validation, got error: %v", err)
	}
}
