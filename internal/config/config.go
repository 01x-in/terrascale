package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const ConfigFileName = "terrascale.yaml"

type Config struct {
	Version    string                `yaml:"version"`
	Project    Project               `yaml:"project"`
	State      StateConfig           `yaml:"state"`
	TenantSpec TenantSpec            `yaml:"tenant_spec"`
	Tiers      map[string]TierPreset `yaml:"tiers,omitempty"`
	Tenants    []Tenant              `yaml:"tenants"`
}

type Project struct {
	Name         string `yaml:"name"`
	TerraformDir string `yaml:"terraform_dir"`
	Mode         string `yaml:"mode"`
	Module       string `yaml:"module,omitempty"`
}

type StateConfig struct {
	Backend       string `yaml:"backend"`
	S3Bucket      string `yaml:"s3_bucket,omitempty"`
	S3Region      string `yaml:"s3_region,omitempty"`
	DynamoDBTable string `yaml:"dynamodb_table,omitempty"`
}

type TierPreset struct {
	VpcMode         string `yaml:"vpc_mode,omitempty"`
	DbInstanceClass string `yaml:"db_instance_class,omitempty"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}
	if err := ValidateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}
	return nil
}

func ValidateConfig(cfg *Config) error {
	if cfg.Version == "" {
		return fmt.Errorf("version is required")
	}
	if cfg.Project.Name == "" {
		return fmt.Errorf("project.name is required")
	}
	if cfg.Project.TerraformDir == "" {
		return fmt.Errorf("project.terraform_dir is required")
	}
	if cfg.Project.Mode != "root" && cfg.Project.Mode != "module" {
		return fmt.Errorf("project.mode must be 'root' or 'module', got %q", cfg.Project.Mode)
	}
	if cfg.Project.Mode == "module" && cfg.Project.Module == "" {
		return fmt.Errorf("project.module is required when mode is 'module'")
	}
	if cfg.State.Backend == "" {
		return fmt.Errorf("state.backend is required")
	}
	if cfg.State.Backend != "local" && cfg.State.Backend != "s3" {
		return fmt.Errorf("state.backend must be 'local' or 's3', got %q", cfg.State.Backend)
	}
	return nil
}

func DefaultConfig(name string) *Config {
	return &Config{
		Version: "1",
		Project: Project{
			Name:         name,
			TerraformDir: ".",
			Mode:         "root",
		},
		State: StateConfig{
			Backend: "local",
		},
		TenantSpec: TenantSpec{
			SharedVariables: make(map[string]string),
		},
		Tiers:   make(map[string]TierPreset),
		Tenants: []Tenant{},
	}
}
