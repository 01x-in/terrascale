package terraform

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateStateDir(t *testing.T) {
	baseDir := t.TempDir()
	slug := "my-tenant"

	stateDir, err := CreateStateDir(baseDir, slug)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := filepath.Join(baseDir, ".terrascale", "state", "my-tenant")
	if stateDir != expected {
		t.Errorf("stateDir = %q, want %q", stateDir, expected)
	}

	info, err := os.Stat(stateDir)
	if err != nil {
		t.Fatalf("state dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory, not file")
	}
}

func TestCreateStateDir_Idempotent(t *testing.T) {
	baseDir := t.TempDir()
	slug := "tenant-1"

	_, err := CreateStateDir(baseDir, slug)
	if err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err = CreateStateDir(baseDir, slug)
	if err != nil {
		t.Fatalf("second create should be idempotent: %v", err)
	}
}

func TestRemoveStateDir(t *testing.T) {
	baseDir := t.TempDir()
	slug := "tenant-to-remove"

	stateDir, _ := CreateStateDir(baseDir, slug)
	// Create a file inside to ensure recursive removal
	if err := os.WriteFile(filepath.Join(stateDir, "terraform.tfstate"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := RemoveStateDir(baseDir, slug); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if StateDirExists(baseDir, slug) {
		t.Error("state dir should be removed")
	}
}

func TestRemoveStateDir_NonExistent(t *testing.T) {
	baseDir := t.TempDir()
	// Should not error when removing non-existent dir
	if err := RemoveStateDir(baseDir, "nonexistent"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStateDirExists(t *testing.T) {
	baseDir := t.TempDir()

	if StateDirExists(baseDir, "nonexistent") {
		t.Error("should not exist")
	}

	CreateStateDir(baseDir, "exists")
	if !StateDirExists(baseDir, "exists") {
		t.Error("should exist")
	}
}

func TestEnsureTerrascaleDir(t *testing.T) {
	baseDir := t.TempDir()
	if err := EnsureTerrascaleDir(baseDir); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dir := filepath.Join(baseDir, ".terrascale", "state")
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("dir not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected directory")
	}
}

func TestGetStatePath(t *testing.T) {
	path := GetStatePath("my-tenant")
	expected := filepath.Join(".terrascale", "state", "my-tenant")
	if path != expected {
		t.Errorf("path = %q, want %q", path, expected)
	}
}
