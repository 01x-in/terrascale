# TerraScale — Product Specification

## Multi-Tenant Lifecycle Manager for Terraform

**Version:** 1.0 (MVP v1) + v2 Roadmap
**Author:** Tushar Sarang
**Last Updated:** March 2026

---

## 1. Problem Statement

Every team that uses Terraform eventually needs to replicate their infrastructure for a second tenant, environment, or client. The journey is always the same:

1. **You write Terraform for one deployment.** Your project has modules for networking, database, storage, compute, and frontend. Together they compose one tenant's infrastructure. It works.

2. **Someone says "we need another one."** A new client signs. QA needs a UAT. Sales needs a demo environment.

3. **You reach for one of three bad options:**
   - **Copy-paste the entire project folder.** Within months you have N copies that have drifted apart. Bug fixes don't propagate. Nobody knows which version is canonical.
   - **Use Terraform workspaces.** Works for two tenants. By tenant four, you've lost track of what's deployed where. One wrong `terraform destroy` in the wrong workspace wrecks your week. No registry, no visibility, no lifecycle management.
   - **Write a custom bash script.** One engineer spends a week building glue code. Only they understand it. They leave. The script becomes untouchable legacy.

4. **All three approaches solve the same fundamental problem:** running the same Terraform project repeatedly with different variables and isolated state. And every team solves it from scratch, poorly, with custom glue code.

**TerraScale is the tool that should have existed.** A lightweight Go CLI that sits on top of any existing Terraform project and handles tenant lifecycle — provisioning, inspection, updates, and decommissioning — with isolated state per tenant.

No AI. No platform. No dashboard. Just a CLI that automates the manual work every team does by hand.

---

## 2. Target User

### Primary: Infrastructure / DevOps Engineers
- Run Terraform daily to manage cloud infrastructure
- Have hit the multi-tenant / multi-environment wall
- Currently managing tenants via copy-paste, workspaces, or scripts
- Want a lightweight tool, not an enterprise platform

### Secondary: Backend / Platform Engineers
- Building SaaS products that need per-client infrastructure isolation
- Responsible for environment provisioning (dev, staging, UAT, demo, production)
- Need to onboard new tenants without manual Terraform work

### Tertiary: Engineering Managers / Tech Leads
- Need visibility into what environments/tenants exist
- Want reproducible, auditable tenant provisioning
- Tired of "only one person knows how to deploy a new environment"

### User Profile
- Comfortable with CLI tools
- Already using Terraform (not evaluating IaC tools)
- Working with AWS, GCP, or Azure (cloud-agnostic, but examples use AWS)
- Team size: 2-20 engineers
- Tenants/environments to manage: 3-50

---

## 3. Core Value Proposition

**One command to provision a fully isolated tenant from your existing Terraform project. No code duplication. No manual state management. No custom scripts.**

### Before TerraScale
```
mkdir infrastructure-tenant-2/
cp -r infrastructure/* infrastructure-tenant-2/
# manually edit variables
# manually configure state backend
# terraform init (hope you got the backend right)
# terraform apply (hope you didn't break tenant-1)
# manually track what exists where
# repeat for tenant 3, 4, 5...
```

### After TerraScale
```
terrascale add city-hospital --var tier=premium --var subdomain=city-hospital
```

---

## 4. Key Insight — How Real Terraform Projects Are Structured

Most Terraform projects are NOT organized as a single "tenant module." They look like this:

```
infrastructure/
├── modules/
│   ├── networking/     # VPC, subnets, security groups
│   ├── ecs/            # ECS cluster, task definitions
│   ├── storage/        # S3 buckets, policies
│   ├── database/       # RDS, parameter groups
│   └── amplify/        # Frontend deployment
├── main.tf             # Composes all modules together
├── variables.tf        # Project-wide variables
├── outputs.tf
└── terraform.tfstate
```

The **entire project IS one tenant.** The `main.tf` composes multiple modules, and together they form one deployment. When teams need a second tenant, they copy this entire structure.

TerraScale treats the **root project** as the tenant unit. It runs your entire project repeatedly with different variables and isolated state. Your modules don't change. Your `main.tf` doesn't change. TerraScale just "stamps" the same infrastructure again with different parameters.

```
Your Terraform Project (the stamp)
  │
  ├── terrascale add parent-hospital  →  state/parent-hospital/  →  VPC 10.0.0.0/16
  ├── terrascale add city-hospital    →  state/city-hospital/    →  VPC 10.1.0.0/16
  ├── terrascale add uat-march        →  state/uat-march/        →  VPC 10.2.0.0/16
  └── terrascale add demo-apollo      →  state/demo-apollo/      →  VPC 10.3.0.0/16
```

Same code. Different variables. Isolated state. That's the product.

---

## 5. Key Features

### MVP v1 — Same Account, Multi-Tenant (Weekend Build)

#### F1: Project Initialization (`terrascale init`)
- Scan existing Terraform project (root-level, not single module)
- Detect all variables from `variables.tf`
- Interactive prompt: select which variables change per tenant vs. stay shared
- Support for both root-project-as-tenant AND single-module-as-tenant modes
- Generate `terrascale.yaml` configuration file
- Create `.terrascale/` directory structure
- Auto-add `.terrascale/state/` to `.gitignore`

#### F2: Tenant Provisioning (`terrascale add <slug>`)
- Validate slug format (lowercase alphanumeric + hyphens, unique)
- Prompt for tenant-specific variables (or accept via `--var` flags)
- Create isolated state directory: `.terrascale/state/<slug>/`
- Generate per-tenant `tenant.tfvars` file
- Generate backend override for state isolation
- Execute: `terraform init` → `terraform plan` → confirm → `terraform apply`
- Capture Terraform outputs and store in registry
- Update tenant status in registry: provisioning → active (or failed)
- Handle partial failures gracefully (mark as failed, preserve state for debugging)

#### F3: Tenant Registry (`terrascale.yaml`)
- YAML-based manifest tracking all tenants
- Per-tenant: slug, name, tier, status, variables, outputs, timestamps, state path
- Machine-readable, human-editable, version-controllable
- Single source of truth for "what tenants exist and where"

#### F4: Tenant Listing (`terrascale list`)
- Table view: slug, name, tier, status, environment, created date
- Filter by: `--status`, `--tier`, `--environment`
- Output formats: table (default), `--json` for scripting

#### F5: Tenant Inspection (`terrascale inspect <slug>`)
- Full detail view: all variables, all outputs, state path, timestamps
- `--outputs-only` flag for piping to other tools
- `--refresh` flag to run `terraform refresh` before showing

#### F6: Tenant Destruction (`terrascale destroy <slug>`)
- Confirmation prompt (type slug to confirm)
- Execute `terraform destroy` scoped to tenant's state
- Clean up state directory
- Update registry (status: destroyed)
- `--auto-approve` for CI/CD usage
- `--keep-state` for audit trail

#### F7: VPC Isolation Tiers
- **Shared VPC mode** (basic tier): Deploy into existing VPC with tenant-specific subnets and security groups. Best for UAT, dev, demos.
- **Dedicated VPC mode** (standard/premium tier): Provision a new VPC per tenant with full network isolation. Best for production clients.
- Tier-driven defaults: basic → always shared, premium → always dedicated, standard → configurable

#### F8: Environment Tagging
- Each tenant tagged with environment type: development, uat, staging, demo, production
- Enables filtering: `terrascale list --environment=uat`
- Informs default configurations (demo gets smaller instances, production gets multi-AZ)

### MVP v2 — Cross-Account Tenancy (2-3 Weeks Post v1)

#### F9: Cross-Account Provisioning
- Optional `aws_account_id` and `aws_role_arn` per tenant
- Generate Terraform provider override with `assume_role` configuration
- Validate IAM permissions before provisioning (`terrascale validate <slug>`)
- Centralized state in management account (state lives where TerraScale runs)

#### F10: Remote State Backend (S3 + DynamoDB)
- Mandatory for cross-account (recommended for same-account team usage)
- Per-tenant state key: `s3://bucket/terrascale/<slug>/terraform.tfstate`
- DynamoDB locking to prevent concurrent modifications

#### F11: Account Management
- `terrascale accounts` — list all AWS accounts and their tenants
- Account-level summary: tenant count, total resources, cost indicators

#### F12: Tenant Migration
- `terrascale migrate <slug>` — move tenant from same-account to cross-account
- Re-provision in target account, migrate state, update registry
- Enables: demo-to-production conversion without rebuilding

---

## 6. Architecture

```
┌──────────────────────────────────────────────────────┐
│                    TerraScale CLI                     │
│   init | add | list | inspect | update | destroy     │
└────────────────────────┬─────────────────────────────┘
                         │
          ┌──────────────┼───────────────────┐
          ▼              ▼                   ▼
┌───────────────┐ ┌────────────────┐ ┌────────────────┐
│ Tenant        │ │ Terraform      │ │ State Backend  │
│ Registry      │ │ Executor       │ │ Manager        │
│               │ │                │ │                │
│ terrascale.   │ │ - Scans .tf    │ │ - Creates dirs │
│ yaml          │ │   files        │ │ - Generates    │
│               │ │ - Generates    │ │   backend      │
│ CRUD ops on   │ │   .tfvars      │ │   overrides    │
│ tenant        │ │ - Runs init/   │ │ - Isolates     │
│ records       │ │   plan/apply/  │ │   state per    │
│               │ │   destroy      │ │   tenant       │
│               │ │ - Captures     │ │ - v2: S3+Dynamo│
│               │ │   outputs      │ │                │
└───────────────┘ └────────────────┘ └────────────────┘
                         │
                         ▼
          ┌──────────────────────────────┐
          │ Your Existing Terraform      │
          │ Project (UNCHANGED)          │
          │                              │
          │ main.tf ← composes modules   │
          │ variables.tf                 │
          │ outputs.tf                   │
          │ modules/                     │
          │   ├── networking/            │
          │   ├── database/              │
          │   ├── storage/               │
          │   ├── ecs/                   │
          │   └── amplify/               │
          └──────────────────────────────┘
```

### How It Works Under The Hood

TerraScale is a wrapper around Terraform. No AI, no magic. When you run `terrascale add city-hospital`:

1. **Creates** `.terrascale/state/city-hospital/` directory
2. **Generates** `tenant.tfvars` with tenant-specific variable values
3. **Generates** backend override pointing state to this tenant's directory
4. **Runs** `terraform init` with backend config for this tenant
5. **Runs** `terraform plan -var-file=tenant.tfvars`
6. **Shows** plan, asks for confirmation
7. **Runs** `terraform apply`
8. **Runs** `terraform output -json` to capture what was created
9. **Saves** everything to `terrascale.yaml` registry

That's it. Generate config, run Terraform three times, save results.

---

## 7. Tech Stack

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | Go 1.22+ | Single binary distribution, no runtime deps. Aligns with infra tooling ecosystem (Terraform, kubectl, Docker CLI). |
| CLI Framework | Cobra | Industry standard for Go CLIs. Subcommands, flags, help generation. |
| Interactive Prompts | Charm (huh + lipgloss) | Modern terminal UI. Beautiful forms and styled output. |
| Table Output | tablewriter | Clean table formatting for `list` command. |
| Config Format | YAML (gopkg.in/yaml.v3) | Human-readable, version-controllable, widely understood. |
| Terraform Interaction | os/exec wrapping `terraform` binary | No Terraform Go SDK dependency. Works with any TF version. |

### Project Structure

```
terrascale/
├── cmd/
│   └── terrascale/
│       └── main.go                 # Entry point
├── internal/
│   ├── cli/
│   │   ├── root.go                 # Root cobra command
│   │   ├── init.go                 # terrascale init
│   │   ├── add.go                  # terrascale add
│   │   ├── list.go                 # terrascale list
│   │   ├── inspect.go              # terrascale inspect
│   │   ├── update.go               # terrascale update (v1 stretch)
│   │   └── destroy.go              # terrascale destroy
│   ├── config/
│   │   ├── config.go               # terrascale.yaml parsing/writing
│   │   ├── tenant.go               # Tenant struct & methods
│   │   └── spec.go                 # TenantSpec & VariableDef
│   ├── terraform/
│   │   ├── executor.go             # Terraform command wrapper
│   │   ├── scanner.go              # .tf file variable discovery
│   │   ├── tfvars.go               # .tfvars file generation
│   │   └── state.go                # State directory management
│   ├── registry/
│   │   └── registry.go             # Tenant CRUD on YAML
│   └── ui/
│       ├── prompt.go               # Interactive prompts (Charm)
│       ├── table.go                # Table output
│       └── spinner.go              # Progress indicators
├── examples/
│   └── demo-project/               # Sample TF project for testing
│       ├── modules/
│       │   ├── networking/
│       │   ├── database/
│       │   ├── storage/
│       │   └── compute/
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── AGENTS.md                        # For Claude Code
└── .goreleaser.yaml
```

---

## 8. Data Structures

### terrascale.yaml (Full Example)

```yaml
version: "1"

project:
  name: "oncology-emr"
  terraform_dir: "."
  mode: "root"                     # "root" (entire project = tenant) or "module" (single module = tenant)
  module: ""                       # only used when mode = "module"

state:
  backend: "local"                 # "local" | "s3"

tenant_spec:
  # Variables that CHANGE per tenant
  tenant_variables:
    - name: "project_name"
      type: "string"
      required: true
      prompt: "Project/tenant identifier"
    - name: "environment"
      type: "string"
      required: true
      default: "production"
      options: ["development", "uat", "staging", "demo", "production"]
    - name: "vpc_cidr"
      type: "string"
      required: true
      prompt: "VPC CIDR block"
      auto_increment: "10.{n}.0.0/16"    # auto-assigns next available CIDR
    - name: "vpc_mode"
      type: "string"
      default: "shared"
      options: ["shared", "dedicated"]
    - name: "shared_vpc_id"
      type: "string"
      required: false
      condition: "vpc_mode == shared"
    - name: "db_instance_class"
      type: "string"
      default: "db.t3.micro"
    - name: "subdomain"
      type: "string"
      required: true

  # Variables that STAY THE SAME across all tenants
  shared_variables:
    aws_region: "ap-south-1"
    ecs_cpu: "512"
    ecs_memory: "1024"
    
  # Terraform outputs to capture per tenant
  outputs:
    - "vpc_id"
    - "db_endpoint"
    - "db_name"
    - "s3_bucket"
    - "api_endpoint"
    - "frontend_url"

# Tier presets — override defaults based on tier
tiers:
  basic:
    vpc_mode: "shared"
    db_instance_class: "db.t3.micro"
  standard:
    vpc_mode: "shared"
    db_instance_class: "db.t3.small"
  premium:
    vpc_mode: "dedicated"
    db_instance_class: "db.t3.medium"

# Managed by terrascale — auto-populated
tenants: []
```

### Tenant Record (in tenants array)

```yaml
tenants:
  - slug: "city-hospital"
    name: "City General Hospital"
    tier: "premium"
    environment: "production"
    status: "active"
    created_at: "2026-03-07T10:30:00Z"
    updated_at: "2026-03-07T10:35:00Z"
    variables:
      project_name: "emr-city-hospital"
      environment: "production"
      vpc_cidr: "10.1.0.0/16"
      vpc_mode: "dedicated"
      db_instance_class: "db.t3.medium"
      subdomain: "city-hospital"
    outputs:
      vpc_id: "vpc-0abc123def456"
      db_endpoint: "emr-city-hospital.cluster-xyz.rds.amazonaws.com"
      db_name: "emr_city_hospital_db"
      s3_bucket: "emr-city-hospital-uploads"
      api_endpoint: "https://api.city-hospital.youremr.com"
      frontend_url: "https://city-hospital.youremr.com"
    state_path: ".terrascale/state/city-hospital/"
    # v2 fields (empty in v1)
    account_mode: "same"
    aws_account_id: ""
    aws_role_arn: ""
```

### Go Structs

```go
// internal/config/config.go

type Config struct {
    Version    string        `yaml:"version"`
    Project    Project       `yaml:"project"`
    State      StateConfig   `yaml:"state"`
    TenantSpec TenantSpec    `yaml:"tenant_spec"`
    Tiers      map[string]TierPreset `yaml:"tiers"`
    Tenants    []Tenant      `yaml:"tenants"`
}

type Project struct {
    Name         string `yaml:"name"`
    TerraformDir string `yaml:"terraform_dir"`
    Mode         string `yaml:"mode"`    // "root" or "module"
    Module       string `yaml:"module"`  // path when mode=module
}

type StateConfig struct {
    Backend       string `yaml:"backend"`
    S3Bucket      string `yaml:"s3_bucket,omitempty"`
    S3Region      string `yaml:"s3_region,omitempty"`
    DynamoDBTable string `yaml:"dynamodb_table,omitempty"`
}

type TenantSpec struct {
    TenantVariables []VariableDef     `yaml:"tenant_variables"`
    SharedVariables map[string]string `yaml:"shared_variables"`
    Outputs         []string          `yaml:"outputs"`
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

type TierPreset struct {
    VpcMode         string `yaml:"vpc_mode,omitempty"`
    DbInstanceClass string `yaml:"db_instance_class,omitempty"`
}

type Tenant struct {
    Slug         string            `yaml:"slug"`
    Name         string            `yaml:"name"`
    Tier         string            `yaml:"tier"`
    Environment  string            `yaml:"environment"`
    Status       TenantStatus      `yaml:"status"`
    CreatedAt    time.Time         `yaml:"created_at"`
    UpdatedAt    time.Time         `yaml:"updated_at"`
    Variables    map[string]string `yaml:"variables"`
    Outputs      map[string]string `yaml:"outputs,omitempty"`
    StatePath    string            `yaml:"state_path"`
    AccountMode  string            `yaml:"account_mode"`
    AWSAccountID string            `yaml:"aws_account_id,omitempty"`
    AWSRoleARN   string            `yaml:"aws_role_arn,omitempty"`
}

type TenantStatus string

const (
    StatusProvisioning TenantStatus = "provisioning"
    StatusActive       TenantStatus = "active"
    StatusUpdating     TenantStatus = "updating"
    StatusDestroying   TenantStatus = "destroying"
    StatusDestroyed    TenantStatus = "destroyed"
    StatusFailed       TenantStatus = "failed"
)
```

```go
// internal/terraform/executor.go

type Executor struct {
    workDir  string
    tfBinary string
}

func (e *Executor) Init(backendConfig map[string]string) error
func (e *Executor) Plan(tfvarsPath string) (*PlanResult, error)
func (e *Executor) Apply(tfvarsPath string, autoApprove bool) (*ApplyResult, error)
func (e *Executor) Destroy(tfvarsPath string, autoApprove bool) error
func (e *Executor) Output() (map[string]string, error)
func (e *Executor) Refresh() error

type PlanResult struct {
    ToAdd     int
    ToChange  int
    ToDestroy int
    RawOutput string
}

type ApplyResult struct {
    Success bool
    Outputs map[string]string
    RawOutput string
}
```

```go
// internal/terraform/scanner.go

type Scanner struct {
    projectDir string
}

// ScanVariables reads all .tf files and extracts variable blocks
func (s *Scanner) ScanVariables() ([]DiscoveredVariable, error)

// ScanModules finds module calls in main.tf
func (s *Scanner) ScanModules() ([]DiscoveredModule, error)

// DetectProjectMode determines if project is root-based or module-based
func (s *Scanner) DetectProjectMode() (string, error)

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
```

---

## 9. Constraints

### Technical Constraints
- **Terraform must be installed.** TerraScale wraps the `terraform` binary via os/exec. It does not embed Terraform or use the Go SDK. User must have `terraform` >= 1.0 on PATH.
- **Local filesystem access.** State directories are created locally. Remote state (S3) is v2.
- **YAML config is the source of truth.** No database. The `terrascale.yaml` file is the single registry. If it's deleted or corrupted, tenant tracking is lost (state files still exist for recovery).
- **Sequential provisioning.** v1 provisions one tenant at a time. No parallel `terraform apply` across tenants.
- **Cloud-agnostic in design, AWS-first in examples.** The tool itself doesn't call AWS APIs. It generates Terraform configs. Examples and tier presets assume AWS (VPC, ECS, RDS) but the core works with any cloud provider.

### Operational Constraints
- **No locking in v1.** Two people running `terrascale add` simultaneously could corrupt the YAML. Team usage requires v2 (remote state + locking).
- **No rollback.** If `terraform apply` fails midway, the tenant is marked "failed." User must manually fix or destroy. Automatic rollback is out of scope.
- **No drift detection in v1.** TerraScale trusts that Terraform state matches reality. `terrascale inspect --refresh` can trigger a state refresh but there's no automated drift alerting.

---

## 10. Out of Scope

### Permanently Out of Scope
- **Web UI / Dashboard.** TerraScale is a CLI tool. Period. If someone wants a UI, they build it on top of TerraScale's YAML registry and JSON output.
- **Replacing Terraform.** TerraScale never generates `.tf` files for your infrastructure. It generates `.tfvars` and backend configs only. Your Terraform code is yours.
- **AI / ML features.** No intelligent provisioning, no cost prediction, no auto-scaling recommendations. It's a deterministic automation tool.
- **Cloud provider API calls.** TerraScale never calls AWS/GCP/Azure APIs directly. All cloud interaction happens through Terraform.

### Out of Scope for v1 (In Scope for v2+)
- Cross-account provisioning (v2)
- Remote state backend S3 + DynamoDB (v2)
- Concurrent tenant provisioning (v2)
- State locking for team usage (v2)
- Tenant migration between accounts (v2)
- Pre/post provisioning hooks (v3)
- Cost estimation per tenant via Infracost (v3)
- RBAC — who can provision/destroy (v3)
- Audit logging (v3)
- `terrascale apply-all` — apply changes across all tenants (v3)

---

## 11. Real-World Usage Scenarios

### Context: Django + React EMR on AWS

**Stack:** Django REST API on ECS, React frontend on AWS Amplify, PostgreSQL on RDS, S3 for uploads, Terraform managing all infrastructure.

### Scenario 1: "QA needs a UAT by tomorrow"

```bash
terrascale add uat-reporting \
  --var environment=uat \
  --var tier=basic \
  --var subdomain=uat-reporting

# 5 minutes later:
# ✓ PostgreSQL database: uat_reporting_db (on shared RDS cluster)
# ✓ S3 bucket: emr-uat-reporting-uploads
# ✓ ECS task with tenant env vars
# ✓ Amplify at uat-reporting.youremr.com
# ✓ Shared VPC (cost-effective for testing)

# When done:
terrascale destroy uat-reporting
```

### Scenario 2: "Sales needs a demo for a prospect"

```bash
terrascale add demo-apollo \
  --var tenant_name="Apollo Hospital Demo" \
  --var environment=demo \
  --var tier=standard \
  --var subdomain=demo-apollo

# Prospect visits demo-apollo.youremr.com
# Demo goes well. They sign. Convert to production:

terrascale update demo-apollo \
  --var tier=premium \
  --var environment=production \
  --var db_instance_class=db.t3.large
```

### Scenario 3: "Second hospital signed"

```bash
terrascale add fortis-hospital \
  --var tenant_name="Fortis Healthcare" \
  --var environment=production \
  --var tier=premium \
  --var subdomain=fortis

# Premium tier → dedicated VPC, dedicated RDS, full isolation
# HIPAA compliance story: each hospital in its own network boundary
```

### Scenario 4: "Show me everything" 

```bash
terrascale list

┌──────────────────┬────────────────────────┬─────────┬────────────┬────────┬─────────────┐
│ SLUG             │ NAME                   │ TIER    │ ENVIRONMENT│ STATUS │ CREATED     │
├──────────────────┼────────────────────────┼─────────┼────────────┼────────┼─────────────┤
│ parent-hospital  │ Parent Company         │ premium │ production │ active │ 2024-01-15  │
│ fortis-hospital  │ Fortis Healthcare      │ premium │ production │ active │ 2026-03-07  │
│ demo-apollo      │ Apollo Hospital Demo   │ standard│ demo       │ active │ 2026-03-05  │
│ uat-reporting    │ UAT Reporting Test     │ basic   │ uat        │ active │ 2026-03-06  │
└──────────────────┴────────────────────────┴─────────┴────────────┴────────┴─────────────┘

terrascale list --environment=production

┌──────────────────┬────────────────────────┬─────────┬────────────┬────────┐
│ SLUG             │ NAME                   │ TIER    │ ENVIRONMENT│ STATUS │
├──────────────────┼────────────────────────┼─────────┼────────────┼────────┤
│ parent-hospital  │ Parent Company         │ premium │ production │ active │
│ fortis-hospital  │ Fortis Healthcare      │ premium │ production │ active │
└──────────────────┴────────────────────────┴─────────┴────────────┴────────┘
```

---

## 12. Build Milestones

### Dependency Graph

```
M1 (Config Layer) ──────┐
                        ├──→ M3 (Init Command) ──→ M5 (Add Command) ──→ M7 (Polish & Ship)
M2 (TF Executor) ───────┘          │                     │
                                   │                     │
                        M4 (Registry) ───────────────────┤
                                                         │
                                   M6 (List/Inspect/Destroy) ──→ M7
```

### Milestone Details

#### M1: Config Layer (PARALLEL — can start immediately)
**Est: 3-4 hours**

Build the YAML configuration parsing and writing layer.

- [ ] Define all Go structs: Config, Project, StateConfig, TenantSpec, VariableDef, TierPreset, Tenant
- [ ] Implement `LoadConfig(path string) (*Config, error)` — parse terrascale.yaml
- [ ] Implement `SaveConfig(config *Config, path string) error` — write terrascale.yaml
- [ ] Implement `ValidateConfig(config *Config) error` — validate required fields
- [ ] Implement slug validation: lowercase, alphanumeric + hyphens, unique
- [ ] Unit tests for parsing, writing, validation

**Output:** `internal/config/` package fully functional and tested.

**No dependencies.** Can start immediately.

---

#### M2: Terraform Executor (PARALLEL — can start immediately)
**Est: 3-4 hours**

Build the Terraform command wrapper.

- [ ] Implement `NewExecutor(workDir, tfBinary string) *Executor`
- [ ] Implement `Init(backendConfig map[string]string) error` — runs `terraform init`
- [ ] Implement `Plan(tfvarsPath string) (*PlanResult, error)` — runs `terraform plan`, parses add/change/destroy counts
- [ ] Implement `Apply(tfvarsPath string, autoApprove bool) (*ApplyResult, error)` — runs `terraform apply`
- [ ] Implement `Destroy(tfvarsPath string, autoApprove bool) error` — runs `terraform destroy`
- [ ] Implement `Output() (map[string]string, error)` — runs `terraform output -json`, parses results
- [ ] Handle stderr/stdout capture and error propagation
- [ ] Detect if `terraform` binary exists on PATH

**Output:** `internal/terraform/executor.go` fully functional and tested.

**No dependencies.** Can start immediately, in parallel with M1.

---

#### M3: Init Command (SEQUENTIAL — depends on M1 + M2)
**Est: 3-4 hours**

Build the `terrascale init` command.

- [ ] Implement `Scanner.ScanVariables()` — parse `variables.tf` to extract variable blocks (name, type, default, description)
- [ ] Implement `Scanner.ScanModules()` — parse `main.tf` for module calls
- [ ] Implement `Scanner.DetectProjectMode()` — determine if root-project or single-module
- [ ] Build interactive prompt flow using Charm `huh`:
  - Detect project structure
  - Ask: "Does this entire project represent one tenant?" (root mode) or "Which module is a tenant?" (module mode)
  - Show discovered variables with checkboxes to mark tenant-specific vs. shared
  - Ask for state backend preference
  - Ask for project name
- [ ] Generate `terrascale.yaml` from user inputs
- [ ] Create `.terrascale/` directory
- [ ] Append `.terrascale/state/` to `.gitignore` (create if doesn't exist)
- [ ] Wire up as `cobra.Command` in `internal/cli/init.go`

**Output:** `terrascale init` works end-to-end.

**Depends on:** M1 (Config structs), M2 (Scanner from TF package)

---

#### M4: Tenant Registry (PARALLEL — can start after M1)
**Est: 2 hours**

Build CRUD operations on the tenants array in terrascale.yaml.

- [ ] Implement `AddTenant(config *Config, tenant Tenant) error` — append to tenants, save
- [ ] Implement `GetTenant(config *Config, slug string) (*Tenant, error)` — find by slug
- [ ] Implement `UpdateTenant(config *Config, slug string, updates map[string]string) error`
- [ ] Implement `UpdateTenantStatus(config *Config, slug string, status TenantStatus) error`
- [ ] Implement `RemoveTenant(config *Config, slug string) error` — mark as destroyed
- [ ] Implement `ListTenants(config *Config, filters TenantFilters) []Tenant`
- [ ] Unit tests for all CRUD operations

**Output:** `internal/registry/registry.go` fully functional.

**Depends on:** M1 (Config structs only — can start as soon as structs are defined)

---

#### M5: Add Command (SEQUENTIAL — core of the product)
**Est: 4-5 hours**

Build the `terrascale add <slug>` command. This is the heart of TerraScale.

- [ ] Parse slug from args, validate format and uniqueness
- [ ] Resolve tier presets (if tier=premium, apply preset vpc_mode=dedicated, etc.)
- [ ] Prompt for required variables (interactive) or parse from `--var` flags
- [ ] Apply auto-increment for variables that support it (e.g., VPC CIDR)
- [ ] Create state directory: `.terrascale/state/<slug>/`
- [ ] Generate `tenant.tfvars` file from collected variables + shared variables
- [ ] Generate backend override file for state isolation
- [ ] Call `executor.Init()` with backend config
- [ ] Call `executor.Plan()` — show plan summary to user
- [ ] Prompt for confirmation (unless `--auto-approve`)
- [ ] Call `executor.Apply()`
- [ ] Call `executor.Output()` — capture outputs
- [ ] Create Tenant record with all data
- [ ] Call `registry.AddTenant()` to persist
- [ ] Handle failure: set status=failed, preserve state for debugging, show error
- [ ] Wire up as `cobra.Command`

**Output:** `terrascale add` works end-to-end — the full provisioning flow.

**Depends on:** M1, M2, M3 (needs working init to have a config), M4 (registry)

---

#### M6: List, Inspect, Destroy Commands (PARALLEL after M4)
**Est: 3-4 hours total**

These are simpler commands that read/modify the registry and optionally call Terraform.

**List (1 hour):**
- [ ] Read config, apply filters (--status, --tier, --environment)
- [ ] Render table using tablewriter
- [ ] Support `--json` output

**Inspect (1 hour):**
- [ ] Read config, find tenant by slug
- [ ] Display all fields in formatted output
- [ ] `--outputs-only` flag
- [ ] `--refresh` flag (calls executor.Refresh())
- [ ] Support `--json` output

**Destroy (1.5 hours):**
- [ ] Confirmation prompt (type slug to confirm)
- [ ] Call `executor.Destroy()` scoped to tenant's state
- [ ] Clean up state directory (or preserve with `--keep-state`)
- [ ] Update registry: status=destroyed
- [ ] `--auto-approve` flag

**Output:** All three commands working.

**Depends on:** M4 (registry). Can run in parallel with M5 for the non-Terraform parts.

---

#### M7: Polish & Ship (SEQUENTIAL — final)
**Est: 3-4 hours**

- [ ] Create demo Terraform project in `examples/demo-project/` (uses local_file + random_password, no real AWS resources)
- [ ] End-to-end test: init → add 3 tenants → list → inspect → destroy one → list again
- [ ] Error messages: make all errors user-friendly with suggested fixes
- [ ] Write README.md with: installation, quick start, full command reference, examples
- [ ] Write AGENTS.md for Claude Code integration
- [ ] Create Makefile: `build`, `test`, `install`, `release`
- [ ] Set up GoReleaser config for binary distribution
- [ ] Push to GitHub
- [ ] Tag v0.1.0

**Output:** Shipped MVP. Public repo. Installable binary.

**Depends on:** All previous milestones.

---

### Parallelization Summary

```
HOUR  0  1  2  3  4  5  6  7  8  9  10 11 12 13 14 15 16
      ├──M1 (Config)──────┤
      ├──M2 (TF Executor)─┤
      │        ├──M4 (Registry)──┤
      │        │                 ├──M3 (Init Cmd)────┤
      │        │                 ├──M6 (List/Ins/Des)┤
      │        │                 │                   ├──M5 (Add Cmd)──────┤
      │        │                 │                   │                   ├──M7 (Polish)───┤
```

**Saturday (8 hrs):** M1 + M2 in parallel → M3 + M4 → start M5
**Sunday (8 hrs):** Finish M5 → M6 → M7 → Ship

---

## 13. Demo Module (For Testing Without AWS Costs)

```hcl
# examples/demo-project/variables.tf

variable "project_name" {
  type        = string
  description = "Tenant project identifier"
}

variable "environment" {
  type        = string
  default     = "production"
  description = "Environment type"
}

variable "tier" {
  type        = string
  default     = "standard"
  description = "Tenant tier"
}

variable "vpc_cidr" {
  type        = string
  default     = "10.0.0.0/16"
  description = "VPC CIDR block"
}

variable "subdomain" {
  type        = string
  description = "Tenant subdomain"
}

variable "db_instance_class" {
  type        = string
  default     = "db.t3.micro"
  description = "Database instance class"
}
```

```hcl
# examples/demo-project/main.tf

resource "local_file" "tenant_config" {
  content = jsonencode({
    project_name     = var.project_name
    environment      = var.environment
    tier             = var.tier
    vpc_cidr         = var.vpc_cidr
    subdomain        = var.subdomain
    db_instance_class = var.db_instance_class
    provisioned_at   = timestamp()
  })
  filename = "${path.module}/.tenant-data/${var.project_name}/config.json"
}

resource "random_password" "api_key" {
  length  = 32
  special = false
}

resource "random_id" "db_suffix" {
  byte_length = 4
}
```

```hcl
# examples/demo-project/outputs.tf

output "vpc_id" {
  value = "vpc-demo-${var.project_name}"
}

output "db_endpoint" {
  value = "${var.project_name}-${random_id.db_suffix.hex}.db.example.com"
}

output "db_name" {
  value = "${replace(var.project_name, "-", "_")}_db"
}

output "s3_bucket" {
  value = "${var.project_name}-uploads"
}

output "api_endpoint" {
  value = "https://api.${var.subdomain}.example.com"
}

output "frontend_url" {
  value = "https://${var.subdomain}.example.com"
}
```

This demo project uses `local_file`, `random_password`, and `random_id` — zero cloud resources, zero cost. Perfect for development and testing the full TerraScale workflow.

---

## 14. Success Criteria

### MVP v1 is "done" when:
1. `terrascale init` scans a real Terraform project and generates valid config
2. `terrascale add` provisions 3 tenants with isolated state (using demo project)
3. `terrascale list` shows all 3 tenants with correct status
4. `terrascale inspect` shows full detail for any tenant
5. `terrascale destroy` cleanly removes one tenant without affecting others
6. `terrascale list` correctly shows 2 active + 1 destroyed
7. The entire flow works without touching any Terraform source files
8. README has clear installation and quick-start instructions
9. Binary builds for Linux, macOS, Windows via GoReleaser

---

## 15. Build Order (For AI Agent / Claude Code)

THIS SECTION IS THE EXPLICIT BUILD PLAN. Follow this exact order. Do not skip ahead. Each phase must compile and pass tests before moving to the next.

### Phase 1: Foundation (No dependencies — build first)

**Step 1A: Go project scaffold**
- Initialize Go module: `github.com/tushar-im/terrascale`
- Set up directory structure as defined in Section 7
- Install dependencies: cobra, yaml.v3, charmbracelet/huh, charmbracelet/lipgloss, tablewriter
- Create `cmd/terrascale/main.go` with root cobra command (no subcommands yet)
- Verify: `go build ./cmd/terrascale/` compiles successfully

**Step 1B: Config layer (internal/config/)**
- Implement ALL Go structs from Section 8: Config, Project, StateConfig, TenantSpec, VariableDef, TierPreset, Tenant, TenantStatus
- Implement `LoadConfig(path) (*Config, error)` — read and parse terrascale.yaml
- Implement `SaveConfig(config, path) error` — marshal and write terrascale.yaml
- Implement `ValidateConfig(config) error` — check required fields, valid values
- Implement `ValidateSlug(slug) error` — lowercase alphanumeric + hyphens only
- Write unit tests for all of the above
- Verify: `go test ./internal/config/...` all pass

**Step 1C: Terraform executor (internal/terraform/executor.go)**
- Implement `NewExecutor(workDir, tfBinary) *Executor`
- Implement `FindTerraformBinary() (string, error)` — locate terraform on PATH
- Implement `Init(backendConfig) error` — shell out to `terraform init`
- Implement `Plan(tfvarsPath) (*PlanResult, error)` — shell out to `terraform plan`, parse output for add/change/destroy counts
- Implement `Apply(tfvarsPath, autoApprove) (*ApplyResult, error)` — shell out to `terraform apply`
- Implement `Destroy(tfvarsPath, autoApprove) error` — shell out to `terraform destroy`
- Implement `Output() (map[string]string, error)` — shell out to `terraform output -json`, parse JSON
- All methods must capture stdout, stderr, and return meaningful errors
- Write unit tests (mock the terraform binary or test with a simple .tf file)
- Verify: `go test ./internal/terraform/...` all pass

### Phase 2: Registry + Scanner (Depends on Phase 1)

**Step 2A: Tenant registry (internal/registry/)**
- Implement `AddTenant(config, tenant) error`
- Implement `GetTenant(config, slug) (*Tenant, error)`
- Implement `UpdateTenant(config, slug, updates) error`
- Implement `UpdateTenantStatus(config, slug, status) error`
- Implement `ListTenants(config, filters) []Tenant`
- Implement `TenantExists(config, slug) bool`
- Write unit tests
- Verify: tests pass

**Step 2B: Terraform scanner (internal/terraform/scanner.go)**
- Implement `ScanVariables(dir) ([]DiscoveredVariable, error)` — parse .tf files for `variable` blocks, extract name, type, default, description
- Implement `ScanModules(dir) ([]DiscoveredModule, error)` — parse .tf files for `module` blocks
- Implement `DetectProjectMode(dir) (string, error)` — returns "root" or "module"
- NOTE: Use simple string/regex parsing of .tf files. Do NOT use the HCL Go library unless it simplifies things significantly.
- Write unit tests using the demo project files from Section 13
- Verify: tests pass

**Step 2C: Tfvars generator (internal/terraform/tfvars.go)**
- Implement `GenerateTfvars(variables map[string]string, path string) error` — write a valid .tfvars file
- Implement `GenerateBackendOverride(stateDir string, path string) error` — write backend config pointing to tenant state dir
- Write unit tests
- Verify: tests pass

**Step 2D: State manager (internal/terraform/state.go)**
- Implement `CreateStateDir(baseDir, slug string) (string, error)` — creates `.terrascale/state/<slug>/`
- Implement `RemoveStateDir(baseDir, slug string) error`
- Implement `StateDirExists(baseDir, slug string) bool`
- Write unit tests
- Verify: tests pass

### Phase 3: CLI Commands (Depends on Phase 2)

Build commands in this exact order. Each command should be fully functional and manually testable before proceeding to the next.

**Step 3A: Create demo project**
- Create `examples/demo-project/` with variables.tf, main.tf, outputs.tf from Section 13
- This is your test fixture for all CLI commands

**Step 3B: `terrascale init` command**
- Wire up cobra subcommand in `internal/cli/init.go`
- Use scanner to discover variables and modules
- Use Charm `huh` for interactive prompts (project mode, variable selection, backend choice)
- Generate terrascale.yaml using config layer
- Create .terrascale/ directory
- Update .gitignore
- Manual test: run `terrascale init` in `examples/demo-project/`, verify terrascale.yaml is correct

**Step 3C: `terrascale add` command — THIS IS THE MOST IMPORTANT COMMAND**
- Wire up cobra subcommand in `internal/cli/add.go`
- Accept slug as positional arg, validate it
- Accept `--var key=value` flags for non-interactive usage
- Resolve tier presets from config
- Prompt for missing required variables (interactive mode)
- Create state directory
- Generate tenant.tfvars
- Generate backend override
- Call executor: init → plan → show plan → confirm → apply
- Capture outputs
- Create tenant record, add to registry
- Handle failures: set status=failed, preserve state, show error
- Manual test: run `terrascale add test-tenant-1` in demo project, verify state dir and registry

**Step 3D: `terrascale list` command**
- Wire up cobra subcommand
- Read config, apply filters (--status, --tier, --environment)
- Render table with tablewriter
- Support --json flag
- Manual test: verify list shows the tenant added in Step 3C

**Step 3E: `terrascale inspect` command**
- Wire up cobra subcommand
- Read config, find tenant by slug
- Display formatted detail view
- Support --outputs-only and --json flags
- Manual test: inspect the tenant from Step 3C

**Step 3F: `terrascale destroy` command**
- Wire up cobra subcommand
- Confirmation prompt (type slug)
- Call executor.Destroy scoped to tenant state
- Update registry status to destroyed
- Clean up state dir (unless --keep-state)
- Support --auto-approve
- Manual test: destroy the tenant, verify list shows status=destroyed

### Phase 4: Polish & Ship (Depends on Phase 3)

**Step 4A: End-to-end verification**
- In demo project: init → add 3 tenants → list → inspect each → destroy one → list
- Verify all state is isolated (destroying one doesn't affect others)
- Fix any bugs found

**Step 4B: Error handling pass**
- Review all commands for user-friendly error messages
- Handle: terrascale.yaml not found, terraform not installed, invalid slug, duplicate slug, state dir already exists, terraform apply failure
- Each error should suggest a fix

**Step 4C: Documentation**
- Write README.md: installation, quick start (5-minute demo), full command reference, examples
- Write AGENTS.md for Claude Code context
- Create Makefile with targets: build, test, install, clean

**Step 4D: Release setup**
- Create .goreleaser.yaml for cross-platform binary builds
- Tag v0.1.0
- Push to GitHub

### CRITICAL RULES FOR THE BUILD:
1. Each Step must compile before moving to the next: `go build ./...`
2. Each Step's tests must pass before moving on: `go test ./...`
3. Do NOT add features not listed in this spec. No "nice to have" additions.
4. Do NOT use Terraform's Go SDK or HCL library unless absolutely necessary. Shell out to the terraform binary.
5. The demo project in examples/ is your primary test environment. Test every command against it.
6. Keep the code simple. This is a weekend MVP, not an enterprise product.

---

*TerraScale — github.com/tushar-im/terrascale*
*Built by Tushar Sarang*
