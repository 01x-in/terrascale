package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type Scanner struct {
	projectDir string
}

type DiscoveredVariable struct {
	Name        string
	Type        string
	Default     string
	Description string
	HasDefault  bool
}

type DiscoveredModule struct {
	Name   string
	Source string
}

func NewScanner(projectDir string) *Scanner {
	return &Scanner{projectDir: projectDir}
}

// ScanVariables reads all .tf files and extracts variable blocks.
func (s *Scanner) ScanVariables() ([]DiscoveredVariable, error) {
	tfFiles, err := filepath.Glob(filepath.Join(s.projectDir, "*.tf"))
	if err != nil {
		return nil, fmt.Errorf("scanning for .tf files: %w", err)
	}
	if len(tfFiles) == 0 {
		return nil, fmt.Errorf("no .tf files found in %s. Make sure you're in a Terraform project directory", s.projectDir)
	}

	var variables []DiscoveredVariable
	for _, tfFile := range tfFiles {
		content, err := os.ReadFile(tfFile)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", tfFile, err)
		}
		vars := parseVariableBlocks(string(content))
		variables = append(variables, vars...)
	}
	return variables, nil
}

// ScanModules finds module calls in .tf files.
func (s *Scanner) ScanModules() ([]DiscoveredModule, error) {
	tfFiles, err := filepath.Glob(filepath.Join(s.projectDir, "*.tf"))
	if err != nil {
		return nil, fmt.Errorf("scanning for .tf files: %w", err)
	}

	var modules []DiscoveredModule
	for _, tfFile := range tfFiles {
		content, err := os.ReadFile(tfFile)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", tfFile, err)
		}
		mods := parseModuleBlocks(string(content))
		modules = append(modules, mods...)
	}
	return modules, nil
}

// DetectProjectMode determines if the project is root-based or module-based.
func (s *Scanner) DetectProjectMode() (string, error) {
	modules, err := s.ScanModules()
	if err != nil {
		return "", err
	}
	// If the project has local module calls, it's a root-based project
	// (the root main.tf composes modules together = one tenant).
	// If no modules, it could be either — default to root.
	if len(modules) > 0 {
		return "root", nil
	}
	return "root", nil
}

// parseVariableBlocks extracts variable blocks from HCL content using regex.
// Matches: variable "name" { ... }
var variableBlockRegex = regexp.MustCompile(`(?s)variable\s+"([^"]+)"\s*\{([^}]*(?:\{[^}]*\}[^}]*)*)\}`)
var typeRegex = regexp.MustCompile(`(?m)^\s*type\s*=\s*(.+?)\s*$`)
var defaultRegex = regexp.MustCompile(`(?m)^\s*default\s*=\s*"([^"]*)"`)
var descriptionRegex = regexp.MustCompile(`(?m)^\s*description\s*=\s*"([^"]*)"`)

func parseVariableBlocks(content string) []DiscoveredVariable {
	var variables []DiscoveredVariable

	matches := variableBlockRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		name := match[1]
		body := match[2]

		v := DiscoveredVariable{Name: name}

		if typeMatch := typeRegex.FindStringSubmatch(body); len(typeMatch) > 1 {
			v.Type = strings.TrimSpace(typeMatch[1])
		}

		if defaultMatch := defaultRegex.FindStringSubmatch(body); len(defaultMatch) > 1 {
			v.Default = defaultMatch[1]
			v.HasDefault = true
		} else if strings.Contains(body, "default") {
			// Handle non-string defaults
			defaultLineRegex := regexp.MustCompile(`(?m)^\s*default\s*=\s*(.+?)\s*$`)
			if dlMatch := defaultLineRegex.FindStringSubmatch(body); len(dlMatch) > 1 {
				val := strings.TrimSpace(dlMatch[1])
				val = strings.Trim(val, `"`)
				v.Default = val
				v.HasDefault = true
			}
		}

		if descMatch := descriptionRegex.FindStringSubmatch(body); len(descMatch) > 1 {
			v.Description = descMatch[1]
		}

		variables = append(variables, v)
	}

	return variables
}

// parseModuleBlocks extracts module blocks from HCL content.
var moduleBlockRegex = regexp.MustCompile(`(?s)module\s+"([^"]+)"\s*\{([^}]*(?:\{[^}]*\}[^}]*)*)\}`)
var sourceRegex = regexp.MustCompile(`(?m)^\s*source\s*=\s*"([^"]*)"`)

func parseModuleBlocks(content string) []DiscoveredModule {
	var modules []DiscoveredModule

	matches := moduleBlockRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		name := match[1]
		body := match[2]

		m := DiscoveredModule{Name: name}
		if sourceMatch := sourceRegex.FindStringSubmatch(body); len(sourceMatch) > 1 {
			m.Source = sourceMatch[1]
		}
		modules = append(modules, m)
	}

	return modules
}
