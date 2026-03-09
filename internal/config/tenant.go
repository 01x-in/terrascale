package config

import (
	"fmt"
	"regexp"
	"time"
)

type TenantStatus string

const (
	StatusProvisioning TenantStatus = "provisioning"
	StatusActive       TenantStatus = "active"
	StatusUpdating     TenantStatus = "updating"
	StatusDestroying   TenantStatus = "destroying"
	StatusDestroyed    TenantStatus = "destroyed"
	StatusFailed       TenantStatus = "failed"
)

type Tenant struct {
	Slug         string            `yaml:"slug"`
	Name         string            `yaml:"name"`
	Environment  string            `yaml:"environment"`
	Status       TenantStatus      `yaml:"status"`
	CreatedAt    time.Time         `yaml:"created_at"`
	UpdatedAt    time.Time         `yaml:"updated_at"`
	Variables    map[string]string `yaml:"variables"`
	Outputs      map[string]string `yaml:"outputs,omitempty"`
	StatePath    string            `yaml:"state_path"`
	AccountMode  string            `yaml:"account_mode"`
	AWSAccountID string            `yaml:"aws_account_id,omitempty"`
	AWSRoleARN   string            `yaml:"aws_role_arn,omitempty"`
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func ValidateSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("slug cannot be empty")
	}
	if len(slug) > 63 {
		return fmt.Errorf("slug must be 63 characters or fewer, got %d", len(slug))
	}
	if !slugRegex.MatchString(slug) {
		return fmt.Errorf("slug %q is invalid: must be lowercase alphanumeric with hyphens (e.g., 'my-tenant-1')", slug)
	}
	return nil
}

func SlugExists(cfg *Config, slug string) bool {
	for _, t := range cfg.Tenants {
		if t.Slug == slug {
			return true
		}
	}
	return false
}
