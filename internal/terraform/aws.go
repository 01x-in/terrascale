package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// AWSIdentity holds the resolved AWS caller identity.
type AWSIdentity struct {
	Profile   string
	AccountID string
	ARN       string
}

// GetAWSIdentity resolves the current AWS profile and caller identity.
// Returns an error if the AWS CLI is not available or credentials are not configured.
func GetAWSIdentity() (*AWSIdentity, error) {
	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = os.Getenv("AWS_DEFAULT_PROFILE")
	}
	if profile == "" {
		profile = "default"
	}

	out, err := exec.Command("aws", "sts", "get-caller-identity", "--output", "json").Output()
	if err != nil {
		return nil, fmt.Errorf("could not resolve AWS identity — check your credentials or AWS_PROFILE: %w", err)
	}

	var result struct {
		Account string `json:"Account"`
		Arn     string `json:"Arn"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parsing AWS identity response: %w", err)
	}

	// Shorten ARN for display: arn:aws:iam::123456789012:user/tushar → user/tushar
	arn := result.Arn
	if parts := strings.SplitN(arn, "::", 2); len(parts) == 2 {
		if sub := strings.SplitN(parts[1], ":", 2); len(sub) == 2 {
			arn = sub[1]
		}
	}

	return &AWSIdentity{
		Profile:   profile,
		AccountID: result.Account,
		ARN:       arn,
	}, nil
}
