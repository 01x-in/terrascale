package config

type TenantSpec struct {
	TenantVariables []VariableDef     `yaml:"tenant_variables"`
	SharedVariables []string          `yaml:"shared_variables,omitempty"`
	Outputs         []string          `yaml:"outputs,omitempty"`
}

type VariableDef struct {
	Name          string   `yaml:"name"`
	Type          string   `yaml:"type"`
	Required      bool     `yaml:"required"`
	Default       string   `yaml:"default,omitempty"`
	Prompt        string   `yaml:"prompt,omitempty"`
	Validation    string   `yaml:"validation,omitempty"`
	Options       []string `yaml:"options,omitempty"`
	AutoIncrement string   `yaml:"auto_increment,omitempty"`
	Condition     string   `yaml:"condition,omitempty"`
}
