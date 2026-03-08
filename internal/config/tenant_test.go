package config

import (
	"strings"
	"testing"
	"time"
)

func TestValidateSlug(t *testing.T) {
	tests := []struct {
		name    string
		slug    string
		wantErr bool
	}{
		// Valid slugs
		{name: "simple slug", slug: "my-tenant", wantErr: false},
		{name: "slug with number", slug: "tenant-1", wantErr: false},
		{name: "single char", slug: "a", wantErr: false},
		{name: "multi-segment", slug: "abc-def-123", wantErr: false},
		{name: "all numbers", slug: "123", wantErr: false},
		{name: "number-letter mix", slug: "1a2b", wantErr: false},
		{name: "exactly 63 chars", slug: strings.Repeat("a", 63), wantErr: false},

		// Invalid slugs
		{name: "empty string", slug: "", wantErr: true},
		{name: "uppercase letters", slug: "My-Tenant", wantErr: true},
		{name: "underscore", slug: "tenant_1", wantErr: true},
		{name: "leading hyphen", slug: "-leading", wantErr: true},
		{name: "trailing hyphen", slug: "trailing-", wantErr: true},
		{name: "double hyphen", slug: "a--b", wantErr: true},
		{name: "over 63 chars", slug: strings.Repeat("a", 64), wantErr: true},
		{name: "spaces", slug: "my tenant", wantErr: true},
		{name: "special chars", slug: "tenant@1", wantErr: true},
		{name: "dots", slug: "tenant.1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSlug(tt.slug)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSlug(%q) error = %v, wantErr %v", tt.slug, err, tt.wantErr)
			}
		})
	}
}

func TestSlugExists(t *testing.T) {
	cfg := &Config{
		Tenants: []Tenant{
			{
				Slug:      "tenant-1",
				Name:      "Tenant One",
				Status:    StatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Variables: map[string]string{},
			},
			{
				Slug:      "tenant-2",
				Name:      "Tenant Two",
				Status:    StatusDestroyed,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Variables: map[string]string{},
			},
		},
	}

	tests := []struct {
		name string
		slug string
		want bool
	}{
		{name: "existing active tenant", slug: "tenant-1", want: true},
		{name: "existing destroyed tenant", slug: "tenant-2", want: true},
		{name: "non-existing tenant", slug: "tenant-3", want: false},
		{name: "empty slug", slug: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SlugExists(cfg, tt.slug)
			if got != tt.want {
				t.Errorf("SlugExists(%q) = %v, want %v", tt.slug, got, tt.want)
			}
		})
	}
}

func TestSlugExists_EmptyConfig(t *testing.T) {
	cfg := &Config{Tenants: []Tenant{}}
	if SlugExists(cfg, "anything") {
		t.Error("SlugExists should return false for empty tenant list")
	}
}
