package terraform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type Executor struct {
	workDir  string
	tfBinary string
}

type PlanResult struct {
	ToAdd     int
	ToChange  int
	ToDestroy int
	RawOutput string
}

type ApplyResult struct {
	Success   bool
	Outputs   map[string]string
	RawOutput string
}

func NewExecutor(workDir, tfBinary string) *Executor {
	return &Executor{
		workDir:  workDir,
		tfBinary: tfBinary,
	}
}

func FindTerraformBinary() (string, error) {
	path, err := exec.LookPath("terraform")
	if err != nil {
		return "", fmt.Errorf("terraform binary not found on PATH. Install Terraform from https://www.terraform.io/downloads")
	}
	return path, nil
}

func (e *Executor) Init(backendConfig map[string]string) error {
	args := []string{"init", "-input=false", "-reconfigure"}
	for key, value := range backendConfig {
		args = append(args, fmt.Sprintf("-backend-config=%s=%s", key, value))
	}
	_, err := e.runStreaming(args...)
	if err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}
	return nil
}

func (e *Executor) Plan(tfvarsPath string) (*PlanResult, error) {
	args := []string{"plan", "-input=false"}
	if tfvarsPath != "" {
		args = append(args, fmt.Sprintf("-var-file=%s", tfvarsPath))
	}
	captured, err := e.runStreaming(args...)
	if err != nil {
		return nil, fmt.Errorf("terraform plan failed: %w", err)
	}
	result := &PlanResult{RawOutput: captured}
	result.ToAdd, result.ToChange, result.ToDestroy = parsePlanCounts(captured)
	return result, nil
}

func (e *Executor) Apply(tfvarsPath string, autoApprove bool) (*ApplyResult, error) {
	args := []string{"apply", "-input=false"}
	if autoApprove {
		args = append(args, "-auto-approve")
	}
	if tfvarsPath != "" {
		args = append(args, fmt.Sprintf("-var-file=%s", tfvarsPath))
	}
	captured, err := e.runStreaming(args...)
	if err != nil {
		return &ApplyResult{Success: false, RawOutput: captured}, fmt.Errorf("terraform apply failed: %w", err)
	}
	return &ApplyResult{Success: true, RawOutput: captured}, nil
}

func (e *Executor) Destroy(tfvarsPath string, autoApprove bool) error {
	args := []string{"destroy", "-input=false"}
	if autoApprove {
		args = append(args, "-auto-approve")
	}
	if tfvarsPath != "" {
		args = append(args, fmt.Sprintf("-var-file=%s", tfvarsPath))
	}
	_, err := e.runStreaming(args...)
	if err != nil {
		return fmt.Errorf("terraform destroy failed: %w", err)
	}
	return nil
}

func (e *Executor) Output() (map[string]string, error) {
	stdout, stderr, err := e.run("output", "-json", "-no-color")
	if err != nil {
		return nil, fmt.Errorf("terraform output failed: %s\n%s", err, stderr)
	}
	return parseOutputJSON(stdout)
}

func (e *Executor) Refresh() error {
	_, stderr, err := e.run("refresh", "-input=false", "-no-color")
	if err != nil {
		return fmt.Errorf("terraform refresh failed: %s\n%s", err, stderr)
	}
	return nil
}

// runStreaming streams stdout/stderr to the terminal and also captures output for parsing.
func (e *Executor) runStreaming(args ...string) (string, error) {
	cmd := exec.Command(e.tfBinary, args...)
	cmd.Dir = e.workDir

	var captured bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &captured)
	cmd.Stderr = io.MultiWriter(os.Stderr, &captured)

	err := cmd.Run()
	return captured.String(), err
}

// run captures output without streaming (used for output -json and refresh).
func (e *Executor) run(args ...string) (string, string, error) {
	cmd := exec.Command(e.tfBinary, args...)
	cmd.Dir = e.workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// ansiRegex strips ANSI escape codes for text parsing.
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// planCountRegex matches "Plan: 3 to add, 1 to change, 0 to destroy."
var planCountRegex = regexp.MustCompile(`Plan:\s+(\d+)\s+to add,\s+(\d+)\s+to change,\s+(\d+)\s+to destroy`)

func parsePlanCounts(output string) (add, change, destroy int) {
	clean := ansiRegex.ReplaceAllString(output, "")
	matches := planCountRegex.FindStringSubmatch(clean)
	if len(matches) == 4 {
		fmt.Sscanf(matches[1], "%d", &add)
		fmt.Sscanf(matches[2], "%d", &change)
		fmt.Sscanf(matches[3], "%d", &destroy)
	}
	return
}

// parseOutputJSON parses terraform output -json and extracts simple string values.
func parseOutputJSON(jsonStr string) (map[string]string, error) {
	if strings.TrimSpace(jsonStr) == "" || strings.TrimSpace(jsonStr) == "{}" {
		return map[string]string{}, nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return nil, fmt.Errorf("parsing terraform output JSON: %w", err)
	}

	result := make(map[string]string, len(raw))
	for key, value := range raw {
		var outputBlock struct {
			Value interface{} `json:"value"`
		}
		if err := json.Unmarshal(value, &outputBlock); err != nil {
			continue
		}
		switch v := outputBlock.Value.(type) {
		case string:
			result[key] = v
		default:
			marshaled, err := json.Marshal(v)
			if err != nil {
				continue
			}
			result[key] = string(marshaled)
		}
	}
	return result, nil
}
