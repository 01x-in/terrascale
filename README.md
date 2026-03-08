# TerraScale

**Multi-tenant lifecycle manager for Terraform.**

TerraScale wraps your existing Terraform project to provision, manage, and destroy isolated tenants — each with their own state file and configuration. No code duplication. No manual state management. No custom scripts.

```
terrascale add city-hospital --var project_name=emr-city --var subdomain=city-hospital
```

One command. Fully isolated tenant. Done.

---

## The Problem

Every team that uses Terraform eventually needs to replicate their infrastructure for a second tenant, environment, or client. The options are all bad:

- **Copy-paste the project folder** — drifts apart within weeks
- **Terraform workspaces** — no registry, no visibility, one wrong `destroy` wrecks everything
- **Custom bash scripts** — only the author understands them

TerraScale solves this. It treats your entire Terraform project as a "stamp" and provisions it repeatedly with different variables and isolated state.

## Installation

### Homebrew (macOS / Linux)

```bash
brew tap 01x-in/tap
brew install terrascale
```

### From Source

```bash
go install github.com/01x-in/terrascale/cmd/terrascale@latest
```

### Build Locally

```bash
git clone https://github.com/01x-in/terrascale.git
cd terrascale
make build
```

### From Release

Download the binary for your platform from the [Releases](https://github.com/01x-in/terrascale/releases) page.

## Quick Start (5 Minutes)

```bash
# 1. Navigate to any existing Terraform project
cd examples/terrascale-site

# 2. Initialize TerraScale
terrascale init

# 3. Add your first tenant
terrascale add tenant-1 \
  --var project_name=test-1 \
  --var subdomain=t1

# 4. Add a second tenant
terrascale add tenant-2 \
  --var project_name=test-2 \
  --var subdomain=t2

# 5. See what you've got
terrascale list

# 6. Inspect a tenant
terrascale inspect tenant-1

# 7. Tear one down (doesn't affect the other)
terrascale destroy tenant-1 --auto-approve

# 8. Confirm isolation
terrascale list
```

## Commands

### `terrascale init`

Scan a Terraform project and generate a `terrascale.yaml` configuration file.

```bash
terrascale init
```

- Discovers all variables from `.tf` files
- Interactive prompts to classify variables as tenant-specific or shared
- Creates `.terrascale/` directory and updates `.gitignore`

### `terrascale add <slug>`

Provision a new tenant with isolated state.

```bash
terrascale add city-hospital \
  --var project_name=emr-city \
  --var subdomain=city-hospital \
  --tier premium \
  --environment production \
  --name "City General Hospital"
```

| Flag | Description | Default |
|------|-------------|---------|
| `--var key=value` | Set a tenant variable (repeatable) | — |
| `--name` | Display name | slug |
| `--tier` | Tier preset: basic, standard, premium | standard |
| `--environment` | Environment type | production |
| `--auto-approve` | Skip confirmation prompt | false |

**What happens under the hood:**
1. Creates `.terrascale/state/<slug>/` for isolated state
2. Generates `tenant.tfvars` with all variables
3. Generates backend override pointing to tenant's state directory
4. Runs `terraform init` → `plan` → `apply`
5. Captures outputs and saves everything to the registry

### `terrascale list`

Display all tenants in a table.

```bash
terrascale list
terrascale list --status=active
terrascale list --tier=premium
terrascale list --environment=production
terrascale list --json
```

| Flag | Description |
|------|-------------|
| `--status` | Filter by status: active, destroyed, failed, provisioning |
| `--tier` | Filter by tier: basic, standard, premium |
| `--environment` | Filter by environment |
| `--json` | Output as JSON |

### `terrascale inspect <slug>`

Show full details for a tenant.

```bash
terrascale inspect city-hospital
terrascale inspect city-hospital --outputs-only
terrascale inspect city-hospital --json
terrascale inspect city-hospital --refresh
```

| Flag | Description |
|------|-------------|
| `--outputs-only` | Show only Terraform outputs |
| `--json` | Output as JSON |
| `--refresh` | Run `terraform refresh` before showing |

### `terrascale destroy <slug>`

Destroy a tenant's infrastructure and clean up.

```bash
terrascale destroy city-hospital
terrascale destroy city-hospital --auto-approve
terrascale destroy city-hospital --keep-state
```

| Flag | Description | Default |
|------|-------------|---------|
| `--auto-approve` | Skip confirmation (for CI/CD) | false |
| `--keep-state` | Preserve state directory for audit | false |

## Configuration

TerraScale uses a single `terrascale.yaml` file as the source of truth. It tracks:

- **Project settings** — Terraform directory, project mode
- **Tenant spec** — which variables change per tenant, which are shared
- **Tier presets** — default variable overrides per tier (basic/standard/premium)
- **Tenant registry** — every tenant's slug, status, variables, outputs, and timestamps

Example:

```yaml
version: "1"
project:
  name: my-infra
  terraform_dir: "."
  mode: root
state:
  backend: local
tenant_spec:
  tenant_variables:
    - name: project_name
      type: string
      required: true
    - name: subdomain
      type: string
      required: true
  shared_variables:
    environment: production
tiers:
  basic:
    vpc_mode: shared
    db_instance_class: db.t3.micro
  premium:
    vpc_mode: dedicated
    db_instance_class: db.t3.medium
tenants: []
```

## How It Works

TerraScale is a wrapper around the `terraform` binary. It never modifies your `.tf` files.

```
Your Terraform Project (the stamp)
  │
  ├── terrascale add hospital-a  →  .terrascale/state/hospital-a/
  ├── terrascale add hospital-b  →  .terrascale/state/hospital-b/
  └── terrascale add uat-march   →  .terrascale/state/uat-march/
```

Same code. Different variables. Isolated state. That's it.

## Requirements

- **Go 1.22+** (to build from source)
- **Terraform >= 1.0** installed and on PATH

## Project Structure

```
terrascale/
├── cmd/terrascale/          # CLI entry point
├── internal/
│   ├── cli/                 # Cobra commands
│   ├── config/              # YAML config structs + I/O
│   ├── terraform/           # TF binary wrapper, scanner, tfvars, state
│   ├── registry/            # Tenant CRUD
│   └── ui/                  # Terminal output helpers
├── examples/
│   └── terrascale-site/     # TerraScale website deployment project
├── terrascale.yaml          # Generated config (per-project)
└── .terrascale/             # Generated state directory (per-project)
```

## Development

```bash
make build       # Build the binary
make test        # Run all tests
make install     # Install to $GOPATH/bin
make lint        # Run linter (requires golangci-lint)
make clean       # Remove built binary
```

## License

MIT
