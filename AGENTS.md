# AGENTS.md

This file provides guidance to AI agents when working with code in this repository.

## Commands

```sh
make build              # Build the provider plugin
make test               # Run unit tests
make testacc            # Run acceptance tests with mocks (default)
make test-update-cassettes  # Run acceptance tests and update cassettes (real API calls)
make fmt                # Format code with gofmt
make vet                # Run go vet
make docs               # Generate documentation
```

To run a single test:
```sh
TF_ACC=1 go test ./internal/services/<service> -v -run=TestName -timeout=120m
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

- Unit tests run with `make test`
- Acceptance tests use VCR cassettes (recorded API interactions) stored in `internal/services/<service>/testdata/`
- Set `TF_UPDATE_CASSETTES=true` to record new cassettes (makes real API calls)
- Some services use VCR v4, others use older versions (see `acctest.go` for details)

### Development Notes

- Many services implemented, each in its own directory under `internal/services/`
- Resources are registered in either `sdkv2.go` or `framework.go` depending on implementation style
- New resources should prefer Framework unless there's a specific reason to use SDKv2
- Linting is configured via `.golangci.yml` with strict formatting and analysis rules
