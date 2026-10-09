# AGENTS.md

This file provides guidance to AI agents when working with code in this repository.

## Commands

```sh
mise install                       # Install all tools pinned in mise.toml
mise tasks                         # List every available task
mise run build:provider            # Build the provider plugin
mise run test:provider             # Run unit tests
mise run test:provider --acceptance            # Run acceptance tests with mocks (default)
mise run test:provider --acceptance --cassettes  # Run acceptance tests and update cassettes (real API calls)
mise run lint:go                   # Lint with golangci-lint
mise run gen:doc                   # Generate documentation
```

To run a single test:
```sh
mise run test:provider --acceptance --run 'TestName' --timeout 120m
```

## Architecture

This is a Terraform Provider for Scaleway implemented in Go. It uses a hybrid architecture combining both `terraform-plugin-sdk/v2` (SDKv2) and `terraform-plugin-framework` (Framework).

### Provider Structure

- **main.go** - Entry point that creates a mux server combining both SDKv2 and Framework providers
- **provider/** - Contains the provider initialization code:
    - `sdkv2.go` - SDKv2 provider with resources and data sources
    - `framework.go` - Framework provider with modern resources/data sources/actions
- **internal/** - Core implementation:
    - `services/` - One directory per Scaleway service (instance, k8s, rdb, etc.), each containing resources, data sources, and testdata with VCR cassettes
    - `meta/` - Shared Meta object with Scaleway SDK client and credentials management
    - `locality/` - Handling of Scaleway zones and regions
    - `transport/` - HTTP transport with retry logic and VCR recording/replay
    - `verify/` - Validation helpers
    - `acctest/` - Acceptance test utilities with VCR mocking
- **cmd/** - Utility commands (vcr-compressor for cassette compression)

### Testing

- Unit tests run with `mise run test:provider`
- Acceptance tests use VCR cassettes (recorded API interactions) stored in `internal/services/<service>/testdata/`; run them with `mise run test:provider --acceptance`
- Record new cassettes (makes real API calls) with `mise run test:provider --acceptance --cassettes`
- Some services use VCR v4, others use older versions (see `acctest.go` for details)

### Development Notes

- Many services implemented, each in its own directory under `internal/services/`
- Resources are registered in either `sdkv2.go` or `framework.go` depending on implementation style
- New resources should prefer Framework unless there's a specific reason to use SDKv2
- Linting is configured via `.golangci.yml` with strict formatting and analysis rules
- Tooling is managed through `mise` (tools and tasks are declared in `mise.toml`); a devcontainer is provided in `.devcontainer/`
