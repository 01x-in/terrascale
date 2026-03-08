package registry

import (
	"fmt"
	"strings"
	"time"

	"github.com/01x-in/terrascale/internal/config"
)

type TenantFilters struct {
	Status      string
	Tier        string
	Environment string
}

func AddTenant(cfg *config.Config, tenant config.Tenant) error {
	if config.SlugExists(cfg, tenant.Slug) {
		return fmt.Errorf("tenant %q already exists. Use a different slug or destroy the existing tenant first", tenant.Slug)
	}
	if err := config.ValidateSlug(tenant.Slug); err != nil {
		return fmt.Errorf("invalid tenant slug: %w", err)
	}
	if tenant.CreatedAt.IsZero() {
		tenant.CreatedAt = time.Now().UTC()
	}
	if tenant.UpdatedAt.IsZero() {
		tenant.UpdatedAt = tenant.CreatedAt
	}
	if tenant.AccountMode == "" {
		tenant.AccountMode = "same"
	}
	cfg.Tenants = append(cfg.Tenants, tenant)
	return nil
}

func GetTenant(cfg *config.Config, slug string) (*config.Tenant, error) {
	for i := range cfg.Tenants {
		if cfg.Tenants[i].Slug == slug {
			return &cfg.Tenants[i], nil
		}
	}
	return nil, fmt.Errorf("tenant %q not found. Run 'terrascale list' to see available tenants", slug)
}

func UpdateTenant(cfg *config.Config, slug string, updates map[string]string) error {
	tenant, err := GetTenant(cfg, slug)
	if err != nil {
		return err
	}
	for key, value := range updates {
		if tenant.Variables == nil {
			tenant.Variables = make(map[string]string)
		}
		tenant.Variables[key] = value
	}
	tenant.UpdatedAt = time.Now().UTC()
	return nil
}

func UpdateTenantStatus(cfg *config.Config, slug string, status config.TenantStatus) error {
	tenant, err := GetTenant(cfg, slug)
	if err != nil {
		return err
	}
	tenant.Status = status
	tenant.UpdatedAt = time.Now().UTC()
	return nil
}

func UpdateTenantOutputs(cfg *config.Config, slug string, outputs map[string]string) error {
	tenant, err := GetTenant(cfg, slug)
	if err != nil {
		return err
	}
	tenant.Outputs = outputs
	tenant.UpdatedAt = time.Now().UTC()
	return nil
}

func TenantExists(cfg *config.Config, slug string) bool {
	return config.SlugExists(cfg, slug)
}

func ListTenants(cfg *config.Config, filters TenantFilters) []config.Tenant {
	var result []config.Tenant
	for _, tenant := range cfg.Tenants {
		if filters.Status != "" && string(tenant.Status) != strings.ToLower(filters.Status) {
			continue
		}
		if filters.Tier != "" && tenant.Tier != strings.ToLower(filters.Tier) {
			continue
		}
		if filters.Environment != "" && tenant.Environment != strings.ToLower(filters.Environment) {
			continue
		}
		result = append(result, tenant)
	}
	return result
}
