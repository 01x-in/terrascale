package registry

import (
	"testing"
	"time"

	"github.com/01x-in/terrascale/internal/config"
)

func testConfig() *config.Config {
	return &config.Config{
		Version: "1",
		Project: config.Project{
			Name:         "test-project",
			TerraformDir: ".",
			Mode:         "root",
		},
		State: config.StateConfig{Backend: "local"},
		Tenants: []config.Tenant{
			{
				Slug:        "tenant-1",
				Name:        "Tenant One",
				Tier:        "basic",
				Environment: "production",
				Status:      config.StatusActive,
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
				Variables:   map[string]string{"project_name": "t1"},
				StatePath:   ".terrascale/state/tenant-1/",
				AccountMode: "same",
			},
			{
				Slug:        "tenant-2",
				Name:        "Tenant Two",
				Tier:        "premium",
				Environment: "uat",
				Status:      config.StatusDestroyed,
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
				Variables:   map[string]string{"project_name": "t2"},
				StatePath:   ".terrascale/state/tenant-2/",
				AccountMode: "same",
			},
		},
	}
}

func TestAddTenant(t *testing.T) {
	cfg := testConfig()
	tenant := config.Tenant{
		Slug:        "tenant-3",
		Name:        "Tenant Three",
		Tier:        "standard",
		Environment: "demo",
		Status:      config.StatusActive,
		Variables:   map[string]string{"project_name": "t3"},
		StatePath:   ".terrascale/state/tenant-3/",
	}

	if err := AddTenant(cfg, tenant); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Tenants) != 3 {
		t.Fatalf("expected 3 tenants, got %d", len(cfg.Tenants))
	}
	if cfg.Tenants[2].AccountMode != "same" {
		t.Errorf("expected default account_mode 'same', got %q", cfg.Tenants[2].AccountMode)
	}
	if cfg.Tenants[2].CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestAddTenant_DuplicateSlug(t *testing.T) {
	cfg := testConfig()
	tenant := config.Tenant{
		Slug: "tenant-1",
		Name: "Duplicate",
	}
	err := AddTenant(cfg, tenant)
	if err == nil {
		t.Fatal("expected error for duplicate slug")
	}
}

func TestAddTenant_InvalidSlug(t *testing.T) {
	cfg := testConfig()
	tenant := config.Tenant{
		Slug: "Invalid_Slug",
		Name: "Bad",
	}
	err := AddTenant(cfg, tenant)
	if err == nil {
		t.Fatal("expected error for invalid slug")
	}
}

func TestGetTenant(t *testing.T) {
	cfg := testConfig()

	tenant, err := GetTenant(cfg, "tenant-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenant.Name != "Tenant One" {
		t.Errorf("expected 'Tenant One', got %q", tenant.Name)
	}
}

func TestGetTenant_NotFound(t *testing.T) {
	cfg := testConfig()
	_, err := GetTenant(cfg, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent tenant")
	}
}

func TestUpdateTenant(t *testing.T) {
	cfg := testConfig()
	err := UpdateTenant(cfg, "tenant-1", map[string]string{
		"project_name": "updated-t1",
		"new_var":      "new-value",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tenant, _ := GetTenant(cfg, "tenant-1")
	if tenant.Variables["project_name"] != "updated-t1" {
		t.Errorf("expected 'updated-t1', got %q", tenant.Variables["project_name"])
	}
	if tenant.Variables["new_var"] != "new-value" {
		t.Errorf("expected 'new-value', got %q", tenant.Variables["new_var"])
	}
}

func TestUpdateTenantStatus(t *testing.T) {
	cfg := testConfig()
	err := UpdateTenantStatus(cfg, "tenant-1", config.StatusDestroyed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tenant, _ := GetTenant(cfg, "tenant-1")
	if tenant.Status != config.StatusDestroyed {
		t.Errorf("expected status 'destroyed', got %q", tenant.Status)
	}
}

func TestUpdateTenantOutputs(t *testing.T) {
	cfg := testConfig()
	outputs := map[string]string{
		"vpc_id":      "vpc-123",
		"db_endpoint": "db.example.com",
	}
	err := UpdateTenantOutputs(cfg, "tenant-1", outputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tenant, _ := GetTenant(cfg, "tenant-1")
	if tenant.Outputs["vpc_id"] != "vpc-123" {
		t.Errorf("expected 'vpc-123', got %q", tenant.Outputs["vpc_id"])
	}
}

func TestTenantExists(t *testing.T) {
	cfg := testConfig()
	if !TenantExists(cfg, "tenant-1") {
		t.Error("expected tenant-1 to exist")
	}
	if TenantExists(cfg, "nonexistent") {
		t.Error("expected nonexistent to not exist")
	}
}

func TestListTenants(t *testing.T) {
	cfg := testConfig()

	tests := []struct {
		name    string
		filters TenantFilters
		want    int
	}{
		{"no filters", TenantFilters{}, 2},
		{"filter by status active", TenantFilters{Status: "active"}, 1},
		{"filter by status destroyed", TenantFilters{Status: "destroyed"}, 1},
		{"filter by tier premium", TenantFilters{Tier: "premium"}, 1},
		{"filter by environment production", TenantFilters{Environment: "production"}, 1},
		{"filter by environment uat", TenantFilters{Environment: "uat"}, 1},
		{"filter no match", TenantFilters{Status: "failed"}, 0},
		{"combined filters", TenantFilters{Status: "active", Tier: "basic"}, 1},
		{"combined filters no match", TenantFilters{Status: "active", Tier: "premium"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ListTenants(cfg, tt.filters)
			if len(result) != tt.want {
				t.Errorf("got %d tenants, want %d", len(result), tt.want)
			}
		})
	}
}
