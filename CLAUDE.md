# CLAUDE.md — TerraScale

## What Is This Project

TerraScale is a Go CLI tool that manages multi-tenant lifecycle on top of any existing Terraform project. It wraps the `terraform` binary to provision, list, inspect, and destroy isolated tenants — each with their own state file and configuration. No AI. No platform. Just Terraform orchestration.

Read `docs/PRODUCT_SPEC.md` for the full product specification, architecture, data structures, and build order. **Follow Section 15 (Build Order) exactly.**

## Project Setup

```bash
go mod tidy
go build ./cmd/terrascale/
```

## Commands

```bash
# Build
go build -o terrascale ./cmd/terrascale/

# Run tests
go test ./... -v

# Run specific package tests
go test ./internal/config/... -v
go test ./internal/terraform/... -v
go test ./internal/registry/... -v

# Lint (if golangci-lint is available)
golangci-lint run ./...

# Run the built binary
./terrascale --help
./terrascale init
./terrascale add <slug>
./terrascale list
./terrascale inspect <slug>
./terrascale destroy <slug>
```

## Build Order — CRITICAL

Follow `PRODUCT_SPEC.md` Section 15 strictly. The phases are:

1. **Phase 1:** Project scaffold → Config layer → Terraform executor
2. **Phase 2:** Registry → Scanner → Tfvars generator → State manager
3. **Phase 3:** Demo project → init cmd → add cmd → list cmd → inspect cmd → destroy cmd
4. **Phase 4:** End-to-end test → Error handling → README → Release config

**Rules:**
- Each step must compile (`go build ./...`) before moving to the next
- Each step's tests must pass (`go test ./...`) before moving on
- **Commit work after each successfully completed task** — every step that compiles and passes tests should be committed before moving on
- Do NOT skip ahead to later phases
- Do NOT add features not in the product spec

## Code Conventions

### Go Style
- Standard Go project layout: `cmd/`, `internal/`, `pkg/`, `examples/`
- Use `internal/` for all packages — nothing is public API in v1 except the CLI
- Error handling: return errors, don't panic. Wrap errors with context using `fmt.Errorf("doing X: %w", err)`
- Keep functions short. If a function is over 50 lines, break it up.
- Use meaningful variable names, not single letters (except loop counters)

### File Organization
- One struct per file when the struct has significant methods
- Group related functions in the same file
- Test files next to source: `config.go` → `config_test.go`

### Dependencies
- **cobra** — CLI framework. All commands go in `internal/cli/`
- **yaml.v3** — Config parsing. Use struct tags.
- **charmbracelet/huh** — Interactive prompts and forms
- **charmbracelet/lipgloss** — Terminal styling (keep it minimal for MVP)
- **tablewriter** — Table output for `list` command
- Do NOT add dependencies beyond these unless absolutely necessary

### Terraform Interaction
- **Always shell out to `terraform` binary via os/exec.** Do NOT use Terraform's Go SDK or the HCL library.
- Capture both stdout and stderr from terraform commands
- Check exit codes — non-zero means failure
- Parse `terraform output -json` for structured output capture
- Parse `terraform plan` stdout for add/change/destroy counts (regex is fine)

### Config File (terrascale.yaml)
- YAML format using gopkg.in/yaml.v3
- Always use `yaml:"field_name"` struct tags
- Use `omitempty` for optional fields
- Load config at command start, save after mutations
- Validate config after loading, before using

### Error Messages
- Every error shown to user must be actionable
- Bad: `"error: config not found"`
- Good: `"terrascale.yaml not found. Run 'terrascale init' first to set up your project."`
- Bad: `"terraform failed"`
- Good: `"terraform apply failed for tenant 'city-hospital'. State has been preserved at .terrascale/state/city-hospital/ for debugging. Run 'terrascale inspect city-hospital' to see current status."`

### Testing
- Use Go's standard `testing` package. No external test frameworks.
- Table-driven tests where appropriate
- Test files live next to source code
- For terraform executor tests: test the tfvars generation and output parsing logic. Don't require terraform binary in unit tests — use integration tests for that.
- The `examples/terrascale-site/` directory is the website deployment example

## File Structure

```
terrascale/
├── cmd/terrascale/main.go          # Entry point
├── internal/
│   ├── cli/                        # Cobra commands
│   │   ├── root.go
│   │   ├── init.go
│   │   ├── add.go
│   │   ├── list.go
│   │   ├── inspect.go
│   │   └── destroy.go
│   ├── config/                     # YAML config structs + I/O
│   │   ├── config.go
│   │   ├── config_test.go
│   │   ├── tenant.go
│   │   ├── tenant_test.go
│   │   └── spec.go
│   ├── terraform/                  # TF binary wrapper
│   │   ├── executor.go
│   │   ├── executor_test.go
│   │   ├── scanner.go
│   │   ├── scanner_test.go
│   │   ├── tfvars.go
│   │   ├── tfvars_test.go
│   │   ├── state.go
│   │   └── state_test.go
│   ├── registry/                   # Tenant CRUD
│   │   ├── registry.go
│   │   └── registry_test.go
│   └── ui/                         # Terminal output helpers
│       ├── prompt.go
│       ├── table.go
│       └── spinner.go
├── examples/
│   └── terrascale-site/            # TerraScale website deployment example
│       ├── main.tf
│       ├── variables.tf
│       └── outputs.tf
├── PRODUCT_SPEC.md                 # Full product specification
├── CLAUDE.md                       # This file
├── README.md
├── go.mod
├── go.sum
├── Makefile
└── .goreleaser.yaml
```

## What NOT To Do

- Do NOT use Terraform Go SDK or HCL parser library — shell out to terraform binary
- Do NOT add a web UI, API server, or database
- Do NOT implement cross-account provisioning (that's v2)
- Do NOT implement remote state / S3 backend (that's v2)
- Do NOT implement state locking (that's v2)
- Do NOT add features beyond what PRODUCT_SPEC.md defines
- Do NOT use external test frameworks (testify, gomega, etc.)
- Do NOT create overly abstract interfaces "for future extensibility" — keep it simple
- Do NOT write long commit messages explaining the architecture — the spec does that

## Verification Checkpoints

After completing each phase, verify:

**After Phase 1:**
```bash
go build ./...                    # compiles
go test ./internal/config/... -v  # config tests pass
go test ./internal/terraform/... -v  # executor tests pass
```

**After Phase 2:**
```bash
go test ./... -v                  # all tests pass
```

**After Phase 3:**
```bash
go build -o terrascale ./cmd/terrascale/
cd examples/terrascale-site
../../terrascale init             # generates terrascale.yaml
../../terrascale add tenant-1 --var project_name=test-1 --var subdomain=t1
../../terrascale add tenant-2 --var project_name=test-2 --var subdomain=t2
../../terrascale list             # shows 2 tenants
../../terrascale inspect tenant-1 # shows detail
../../terrascale destroy tenant-1 --auto-approve
../../terrascale list             # shows 1 active, 1 destroyed
```

**After Phase 4:**
```bash
go test ./... -v                  # all tests still pass
goreleaser build --snapshot       # builds binaries (if goreleaser installed)
```
