package terraform

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	TerrascaleDir = ".terrascale"
	StateDir      = "state"
)

// CreateStateDir creates the state directory for a tenant at .terrascale/state/<slug>/.
func CreateStateDir(baseDir, slug string) (string, error) {
	stateDir := filepath.Join(baseDir, TerrascaleDir, StateDir, slug)
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return "", fmt.Errorf("creating state directory for tenant %q: %w", slug, err)
	}
	return stateDir, nil
}

// RemoveStateDir removes the state directory for a tenant.
func RemoveStateDir(baseDir, slug string) error {
	stateDir := filepath.Join(baseDir, TerrascaleDir, StateDir, slug)
	if err := os.RemoveAll(stateDir); err != nil {
		return fmt.Errorf("removing state directory for tenant %q: %w", slug, err)
	}
	return nil
}

// StateDirExists checks if a state directory exists for a tenant.
func StateDirExists(baseDir, slug string) bool {
	stateDir := filepath.Join(baseDir, TerrascaleDir, StateDir, slug)
	info, err := os.Stat(stateDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// EnsureTerrascaleDir creates the .terrascale/ directory structure.
func EnsureTerrascaleDir(baseDir string) error {
	dir := filepath.Join(baseDir, TerrascaleDir, StateDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating .terrascale directory: %w", err)
	}
	return nil
}

// GetStatePath returns the relative state path for a tenant.
func GetStatePath(slug string) string {
	return filepath.Join(TerrascaleDir, StateDir, slug)
}
