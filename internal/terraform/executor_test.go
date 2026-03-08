package terraform

import (
	"testing"
)

func TestParsePlanCounts(t *testing.T) {
	tests := []struct {
		name              string
		output            string
		wantAdd, wantChange, wantDestroy int
	}{
		{
			name:        "standard plan output",
			output:      "Plan: 3 to add, 1 to change, 0 to destroy.",
			wantAdd:     3,
			wantChange:  1,
			wantDestroy: 0,
		},
		{
			name:        "all zeros",
			output:      "Plan: 0 to add, 0 to change, 0 to destroy.",
			wantAdd:     0,
			wantChange:  0,
			wantDestroy: 0,
		},
		{
			name:        "large numbers",
			output:      "Plan: 15 to add, 7 to change, 3 to destroy.",
			wantAdd:     15,
			wantChange:  7,
			wantDestroy: 3,
		},
		{
			name:        "embedded in larger output",
			output:      "Some preamble\n\nPlan: 2 to add, 0 to change, 1 to destroy.\n\nMore text",
			wantAdd:     2,
			wantChange:  0,
			wantDestroy: 1,
		},
		{
			name:        "no match",
			output:      "No changes. Infrastructure is up-to-date.",
			wantAdd:     0,
			wantChange:  0,
			wantDestroy: 0,
		},
		{
			name:        "empty output",
			output:      "",
			wantAdd:     0,
			wantChange:  0,
			wantDestroy: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			add, change, destroy := parsePlanCounts(tt.output)
			if add != tt.wantAdd {
				t.Errorf("add = %d, want %d", add, tt.wantAdd)
			}
			if change != tt.wantChange {
				t.Errorf("change = %d, want %d", change, tt.wantChange)
			}
			if destroy != tt.wantDestroy {
				t.Errorf("destroy = %d, want %d", destroy, tt.wantDestroy)
			}
		})
	}
}

func TestParseOutputJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "empty object",
			input: "{}",
			want:  map[string]string{},
		},
		{
			name:  "empty string",
			input: "",
			want:  map[string]string{},
		},
		{
			name: "string outputs",
			input: `{
				"vpc_id": {"value": "vpc-123", "type": "string"},
				"db_endpoint": {"value": "db.example.com", "type": "string"}
			}`,
			want: map[string]string{
				"vpc_id":      "vpc-123",
				"db_endpoint": "db.example.com",
			},
		},
		{
			name: "numeric output",
			input: `{
				"port": {"value": 5432, "type": "number"}
			}`,
			want: map[string]string{
				"port": "5432",
			},
		},
		{
			name:    "invalid json",
			input:   "not json",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOutputJSON(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("got %d outputs, want %d", len(got), len(tt.want))
				return
			}
			for key, wantVal := range tt.want {
				if gotVal, ok := got[key]; !ok {
					t.Errorf("missing key %q", key)
				} else if gotVal != wantVal {
					t.Errorf("key %q = %q, want %q", key, gotVal, wantVal)
				}
			}
		})
	}
}

func TestNewExecutor(t *testing.T) {
	exec := NewExecutor("/tmp/test", "/usr/bin/terraform")
	if exec.workDir != "/tmp/test" {
		t.Errorf("workDir = %q, want /tmp/test", exec.workDir)
	}
	if exec.tfBinary != "/usr/bin/terraform" {
		t.Errorf("tfBinary = %q, want /usr/bin/terraform", exec.tfBinary)
	}
}
