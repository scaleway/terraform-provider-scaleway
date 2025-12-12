# Agent Development Guide

This document provides guidelines for developers working on the Scaleway Terraform provider. It covers the project architecture, coding standards, and best practices for contributing to the codebase.

## Development Environment Setup

Before contributing to the Scaleway Terraform provider, you need to set up your development environment with the following requirements:

### Go Toolchain

- Install Go with a version that matches or exceeds the version specified in `go.mod` (currently Go 1.25.0)
- The Go toolchain is required for building, testing, and running the provider

### Scaleway Credentials

To run acceptance tests and interact with the Scaleway API, you'll need credentials. The preferred method is using the Scaleway configuration file:

- **Preferred method**: Configure credentials in the Scaleway configuration file at `~/.config/scw/config.yaml`
  - This is the recommended approach as it keeps credentials secure and persistent across sessions
  - The configuration file can be created and managed using the Scaleway CLI
  - It allows for easier management of multiple profiles and regions/zones

- Alternative method: Set credentials via environment variables:
  - `SCW_ACCESS_KEY`: Your Scaleway access key
  - `SCW_SECRET_KEY`: Your Scaleway secret key
  - `SCW_DEFAULT_REGION`: Default region (e.g., fr-par)
  - `SCW_DEFAULT_ZONE`: Default zone (e.g., fr-par-1)

- If no configuration file exists, use `scw login` (if the Scaleway CLI is installed)
  - This command authenticates you and creates the configuration file
  - The CLI will prompt for your access key and secret key
  - It will also allow you to set default region and zone

Using the configuration file is preferred because it:
- Keeps credentials secure and out of your shell environment
- Persists across terminal sessions
- Allows for easier management of multiple profiles
- Is consistent with other Scaleway tools and documentation

### Required Tools

Install the following tools to work on the provider:

#### Typos Checker
- Install the [`typos`](https://github.com/crate-ci/typos) tool to check for typos in code and documentation
- Used in the CI pipeline for spell checking

#### golangci-lint
- Install `golangci-lint` with the version specified in the lint workflow (v2.5.0)
- This ensures consistency with the CI/CD pipeline
- The linter configuration is defined in `.golangci.yml`

#### Make
- Ensure `make` is installed on your system
- The project uses Makefiles for various development tasks (build, test, lint, etc.)

Once your environment is set up, you can proceed with development following the guidelines in the subsequent sections.

## Project Architecture

The Scaleway Terraform provider follows a modular architecture with the following key components:

- **`internal/services/`**: Contains individual service implementations (e.g., `account`, `instance`, `rdb`). Each service directory includes:
  - Resource and data source implementations
    - Acceptance tests with VCR cassettes in `testdata/` (automatically recorded by tests)
    - Service-specific clients and helpers

- **`provider/`**: Contains the provider implementation using both Terraform Plugin Framework and SDKv2

- **`internal/locality/`**: Handles resource localization (zone/region) for IDs

- **`cmd/`**: Contains utility commands like `vcr-compressor` for cassette management

## Adding New Resources

When adding a new resource to the provider, follow this checklist to ensure consistency and proper implementation:

### TODO List for Adding a New Resource

1. **Create the resource file**
   - Add a new Go file in the appropriate service directory (e.g., `internal/services/{service}/resource_{name}.go`)
   - Implement the resource schema using `schema.Resource`
   - Follow existing naming conventions

2. **Implement CRUD operations**
   - Create: `resource{Service}{Resource}Create`
   - Read: `resource{Service}{Resource}Read`
   - Update: `resource{Service}{Resource}Update`
   - Delete: `resource{Service}{Resource}Delete`
   - Always return early with diagnostics when error codes are not nil
   - Include comprehensive context in diagnostic messages about the request and resource

3. **Add schema definition**
   - Define required, optional, and computed attributes
   - Include proper validation functions
   - Add descriptive field descriptions

4. **Implement data source (if applicable)**
   - Create corresponding data source file
   - Implement read functionality
   - Define appropriate filters

5. **Write acceptance tests**
   - Create test file in the service directory
   - Use the `acctest` package for test helpers
   - Ensure tests are isolated and can run in parallel
   - Test both creation and update scenarios

6. **Handle localization properly**
   - Use the locality package for zone/region handling
   - Ensure IDs follow the proper format (see ID Format section below)
   - Use appropriate parsing functions

7. **Register the resource**
   - If using Terraform Plugin Framework: Add the resource to the Resources() method in `provider/framework.go`
   - If using Terraform SDKv2: Add the resource to the ResourcesMap in `provider/sdkv2.go`
   - Follow the naming convention: `scaleway_{service}_{resource_name}`
   - Ensure the resource function is properly imported

8. **Add documentation**
   - Create a new Go template file (`.tmpl`) in the appropriate subdirectory of `templates/` (e.g., `templates/resources/` for resources)
   - Ensure examples are provided in the template
   - Document all parameters and attributes in the template
   - Run `make docs` to generate the documentation from the templates

9. **Run validation**
   - Execute `make build` to ensure the code compiles successfully
   - Execute `make lint` to ensure the code passes all linter checks
   - Execute `make testacc` to verify acceptance tests pass
   - Execute `make format` to ensure proper code formatting

## Testing Guidelines

This section outlines the testing practices and guidelines for the Scaleway Terraform provider. Additional details can be found in the TESTING.md file, which should be followed when implementing tests.

### Unit Testing

- Follow the guidance in the official Terraform documentation for unit testing: [Unit Testing](https://developer.hashicorp.com/terraform/plugin/sdkv2/testing/unit-testing)
- Unit tests should focus on logic and validation without external dependencies
- Do not use mocking in unit tests - rely on pure function testing

### Acceptance Testing

- Only use acceptance testing with VCR cassettes to record actual API interactions
- Do not use mocking - cassettes provide more reliable and accurate testing
- Acceptance tests for a service should be in a dedicated package named after the service with `_test` suffix
- Follow the standards in the official documentation: [Acceptance Tests](https://developer.hashicorp.com/terraform/plugin/sdkv2/testing/acceptance-tests)

### Test Structure and Validation

- In test steps, verify important attributes of the schema as described in: [TestStep](https://developer.hashicorp.com/terraform/plugin/sdkv2/testing/acceptance-tests/teststep)
- Include steps that modify configuration to verify the update functionality works properly
- Test both creation and subsequent updates in the same test flow
- Ensure tests are isolated and can run in parallel

## Cassette Management Guardrails

VCR cassettes (recorded API interactions) are essential for reliable acceptance testing but can become very large. To prevent context overload:

### Best Practices

1. **Never Manually Modify Cassettes**
   - Cassettes are automatically recorded by acceptance tests and should never be manually modified
   - Always run tests with `TF_UPDATE_CASSETTES=true` to properly record new interactions
   - Manual modifications can break test consistency and lead to unreliable tests

2. **Cassette Compression**
   - Always compress cassettes using the vcr-compressor tool:
   ```bash
   go run ./cmd/vcr-compressor internal/services/{service}/testdata/{resource}.cassette
   ```
   - Compressed cassettes are automatically detected and decompressed during testing

3. **Selective Cassette Loading**
   - Never load all cassettes by default in any context
   - Load only the specific cassette needed for the current operation
   - Use pattern matching to load cassettes by service or resource type when necessary

4. **Cassette Organization**
   - Store cassettes in `internal/services/{service}/testdata/`
   - Use descriptive names (e.g., `{resource}-basic.cassette.yaml`)
   - Group related cassettes in the same directory

5. **Test Isolation**
   - Each test should use its own cassette when possible
   - Avoid sharing cassettes between unrelated tests
   - Use unique test prefixes to prevent interference

6. **Update Strategy**
   - Only update cassettes when API changes require it
   - Set `TF_UPDATE_CASSETTES=true` explicitly when needed
   - Review cassette changes carefully in pull requests

## Coding Style and Linting

The project uses `golangci-lint` for code quality enforcement with a comprehensive configuration.

### Linting Optimization

To avoid long linting times on the entire codebase:

1. **Run linting on changed files only**
   ```bash
   # Run on specific files/directories
   golangci-lint run internal/services/account/...

   # Run on staged files only (recommended for pre-commit)
   git diff --name-only --cached | grep '\.go$' | xargs golangci-lint run
   ```

2. **Linter Configuration**
   - Configuration is defined in `.golangci.yml`
   - The file enables 60+ linters including formatting, performance, and security checks
   - Some linters are disabled (cyclop, dupl, etc.) based on project needs

3. **Common Linters in Use**
   - `gofmt`/`goimports`: Code formatting and import ordering
   - `errcheck`: Ensures errors are handled
   - `staticcheck`: Comprehensive static analysis
   - `gocyclo`: Cyclomatic complexity checking
   - `wsl`: Whitespace validation

## Terraform Resource ID Format

Resource IDs must follow a specific locality format to properly identify resources within the Scaleway infrastructure.

### ID Structure

Most resources IDs must include the zone or region as a prefix, followed by the actual resource ID:

```
{locality}/{resource-id}
```

### Format Examples

**Zonal Resources** (specific availability zone):
```
fr-par-1/d1f93913-8b8e-4883-828e-0d85e743d48c
nl-ams-1/2a3b4c5d-6e7f-8g9h-0i1j-2k3l4m5n6o7p
```

**Regional Resources** (entire region):
```
fr-par/11111111-1111-1111-1111-111111111111
nl-ams/22222222-2222-2222-2222-222222222222
```

### Implementation Guidelines

1. **Use the locality package**
   - Import `github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal` or `regional`
   - Use `zonal.ParseID()` or `regional.ParseID()` to extract zone/region and ID
   - Use `zonal.NewIDString()` or `regional.NewIDString()` to construct IDs

2. **Handle ID parsing in CRUD operations**
   ```go
   // Example of parsing a zonal ID
   zone, id, err := zonal.ParseID(d.Id())
   if err != nil {
       return diag.FromErr(err)
   }

   // Use zone and id in API calls
   res, err := api.GetResource(&api.GetResourceRequest{
       Zone:     zone,
       ResourceID: id,
   })
   ```

3. **Set proper IDs in state**
   ```go
   // When creating a resource, set the full zonal/regional ID
   d.SetId(zonal.NewIDString(zone, res.ID))
   ```

4. **Validation**
   - The parsing functions will validate the format
   - Ensure error handling for malformed IDs
   - Use the `ExpandID` function when you need just the resource ID

## Documentation

### Official Documentation
- **Provider Documentation**: [https://registry.terraform.io/providers/scaleway/scaleway/latest/docs](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs)
- **Terraform Provider Reference**: [https://www.terraform.io/docs/providers/scaleway/index.html](https://www.terraform.io/docs/providers/scaleway/index.html)
- **Scaleway Platform Documentation**: [https://www.scaleway.com/en/docs/](https://www.scaleway.com/en/docs/)

### Local Documentation
- `README.md`: Project overview and setup instructions
- `TESTING.md`: Testing guidelines and procedures
- Service-specific documentation in `docs/` directory
- Code comments and Go documentation
- Template files in `templates/` directory (Go templates used to generate documentation for resources)

The `docs` folder is generated using the `templates` folder which contains Go templates (with `.tmpl` extension) that are rendered to produce the final documentation. When adding documentation for a new resource or data source:
1. Create the appropriate `.tmpl` file in the corresponding subdirectory of `templates/` (e.g., `templates/resources/` for resources, `templates/data-sources/` for data sources)
2. Ensure the template follows the existing patterns and includes all necessary sections (examples, arguments, attributes)
3. Run `make docs` to generate the documentation from the templates

## Additional Resources

- **Slack Community**:
  - Scaleway Community Slack: [https://slack.scaleway.com/](https://slack.scaleway.com/)
  - Terraform Channel: `#terraform`

- **Issue Tracking**: GitHub Issues for bug reports and feature requests
  - For bug reports, please use the bug report template which can be found at `.github/ISSUE_TEMPLATE/bug-report.md`
  - The bug report template provides the necessary structure to include all relevant information for efficient troubleshooting

## Release Process

The release process is automated using GitHub Actions and is triggered when a new tag is pushed to the repository. The process is defined in `.github/workflows/release.yml` and follows these steps:

### Version Management and Release Steps

1. **Fetch Latest Tags**: Before starting, ensure you have all the latest tags from the remote repository:
   ```bash
   git fetch --tags
   ```

2. **Identify Previous Version**: Find the most recent release tag:
   ```bash
   git describe --tags --abbrev=0
   ```
   This will return the latest tag (e.g., `v2.61.0`).

3. **Analyze Changes**: Examine the changes between the current state and the previous tag to determine the appropriate version increment:
   ```bash
   git log --oneline <previous-tag>..HEAD
   ```
   - If the changes include **new features or significant enhancements**, increment the **minor version** (e.g., `v2.61.0` → `v2.62.0`)
   - If the changes are **bug fixes or minor improvements**, increment the **patch version** (e.g., `v2.61.0` → `v2.61.1`)

4. **Create and Push New Tag**: Once you've determined the correct version increment:
   ```bash
   git tag v<x>.<y>.<z>
   git push origin v<x>.<y>.<z>
   ```
   Replace `<x>.<y>.<z>` with the appropriate version number.

5. **Release Automation**: The GitHub Action will automatically:
   - Checkout the code at the tagged commit
   - Install the appropriate version of Go
   - Verify that `go.mod` is properly formatted
   - Import the GPG key for signing (configured via GitHub secrets)
   - Use GoReleaser to build and package binaries for multiple platforms
   - Create a new GitHub release with the generated artifacts

### Versioning Guidelines

- The version format follows semantic versioning: `v<major>.<minor>.<patch>`
- The version in the codebase (defined in `version/version.go`) is automatically updated by GoReleaser during the release process
- Always review the `CHANGELOG.md` to understand the scope of changes before deciding on version increment
- For significant breaking changes, consult the team before incrementing the major version

### Example Workflow

```bash
# 1. Fetch all remote tags
git fetch --tags

# 2. Find the latest tag
git describe --tags --abbrev=0
# Returns: v2.61.0

# 3. Review changes since last release
git log --oneline v2.61.0..HEAD

# 4. If changes include new features, increment minor version
git tag v2.62.0
git push origin v2.62.0
```

The release process uses GoReleaser (configured in `.goreleaser.yml`) to handle the multi-platform builds and packaging. The GPG key for signing is securely stored in GitHub secrets (`GPG_PRIVATE_KEY` with passphrase in `GPG_PASSPHRASE`).

## Dependency Management

The project uses Go modules for dependency management, with dependencies declared in `go.mod` without a vendor directory.

### Dependency Updates
- External dependencies are updated regularly using Dependabot, which is configured in the `.github` repository
- The Scaleway Go SDK is upgraded regularly on the master branch to ensure the latest version is available
- Security updates are automatically provided through Dependabot

### Dependency Management Process
1. **Automatic Updates**: Dependabot creates pull requests for dependency updates, which are then reviewed and merged
2. **Manual Updates**: When adding or updating dependencies manually:
   - Use `go get` to add or update a dependency
   - Run `go mod tidy` to clean up the module file and download dependencies
   - Verify that the changes work correctly by running `make build` and `make testacc`

3. **Module Verification**: Always ensure that `go.mod` and `go.sum` files are committed together with any code changes that involve new or updated dependencies.

By following these practices, we ensure that the provider always uses compatible and secure versions of its dependencies while minimizing manual intervention in the update process.

## Code Organization Conventions

The project follows Go best practices for code organization, with specific conventions to ensure consistency and maintainability across the codebase.

### Package Naming

- Avoid ambiguous package names such as `util`, `utils`, `helper`, or `common`
- Use descriptive package names that clearly indicate their purpose and scope
- Follow the Single Responsibility Principle - each package should have a clear, focused responsibility
- Package names should be lowercase with no underscores or hyphens

### Go Standards and Conventions

The codebase adheres to the guidelines outlined in the official Go documentation, particularly the "Effective Go" guide:

- Follow the recommendations in [Effective Go](https://go.dev/doc/effective_go) for all aspects of code organization
- Maintain consistency in code layout, formatting, and structure
- Use Go idioms and patterns as described in the documentation

### Documentation Conventions

- Use GoDoc style comments for all public functions, types, and variables
- Write clear, concise comments that explain the "why" not just the "what"
- Include examples in documentation when appropriate
- Keep comments up-to-date with code changes

### Naming Conventions

- Use clear, descriptive names for packages, functions, variables, and types
- Follow Go naming conventions:
  - `MixedCaps` for exported identifiers
  - `mixedCaps` for unexported identifiers
  - Use `ID` instead of `Id` or `id` in names
- Choose variable names that reflect their purpose and usage
- Use plural names for packages that contain multiple related types or functions

### Code Layout

- Group related code together and separate unrelated functionality
- Use blank lines to separate logical sections within files
- Order declarations in a logical sequence (typically: constants, variables, types, functions)
- Keep functions focused and limited in scope

### Best Practices

- Keep packages focused and cohesive
- Minimize package dependencies to reduce coupling
- Use Go's standard library when possible instead of external dependencies
- Write testable code by separating concerns and minimizing side effects
- Follow the principle of least surprise in API design

By adhering to these code organization conventions, we ensure that the codebase remains maintainable, readable, and consistent as it continues to grow and evolve.

## Contribution Workflow

The project follows a standard GitHub flow for contributions, with specific guidelines to ensure code quality and maintainability.

### Development Process

- Use the standard GitHub flow: create a branch, commit changes, push to GitHub, and open a pull request
- All changes must go through pull requests and receive approval before merging
- Branch names should be descriptive and relate to the feature or fix being implemented

### Pull Request Guidelines

- Pull requests should be focused on a single feature or fix
- Each PR should concern only one resource when possible
- Keep PRs small and focused to facilitate review
- Include relevant test updates and documentation changes
- Reference related issues in the PR description

### Breaking Changes

- Avoid breaking changes whenever possible
- When a breaking change is unavoidable:
  - Follow the guidance in the official Terraform documentation: [Deprecations](https://developer.hashicorp.com/terraform/plugin/framework/deprecations)
  - Maintain compatibility within the current major version (v2)
  - Provide clear migration paths for users
  - Document changes thoroughly in the PR and release notes

### Deprecation Process

When deprecating functionality:
- Follow Hashicorp's official documentation for deprecations
- Provide clear warnings to users about upcoming changes
- Maintain backward compatibility for a reasonable period
- Document the replacement approach and migration steps

### Review Process

- All PRs require at least one approval from a maintainer
- Address all feedback and requested changes before merging
- Update PRs to resolve any merge conflicts
- Ensure all CI checks are passing before requesting review

By following this contribution workflow, we maintain a high standard of code quality and ensure a smooth development process for all contributors.

## Security Practices

When developing resources for the Scaleway Terraform provider, security considerations are paramount. Follow these guidelines to ensure sensitive data is properly handled:

### Sensitive Fields

- All fields that contain sensitive data (such as passwords, certificates, private keys, API tokens, etc.) must be marked as sensitive in the schema
- Use the `sensitive: true` attribute in the schema definition:
  ```go
  &schema.Schema{
      Type:        schema.TypeString,
      Sensitive:   true,
      Description: "Password for the resource (will be marked as sensitive)",
  }
  ```

### Write-Only Fields

- When possible, provide write-only field options following Hashicorp's guidance: [Write-Only Arguments](https://developer.hashicorp.com/terraform/plugin/sdkv2/resources/write-only-arguments)

### Testing Sensitive Data

- When writing acceptance tests that require passwords or other sensitive values, use dummy values.
- Never use real credentials or passwords in tests
- Ensure dummy values meet the API's validation requirements (length, character types, etc.)

### Documentation Requirements

- Any sensitive field must be clearly documented as such in the attribute description
- Include security warnings in the resource documentation template (.tmpl file)
- Explain why the field is sensitive and any security implications for users
- Document whether the field is write-only (if applicable)

### Implementation Guidelines

1. **Schema Definition**: Always set `sensitive: true` for appropriate fields
2. **State Management**: Sensitive data will be obscured in state files and command output
3. **Error Handling**: Ensure error messages don't leak sensitive information
4. **Logging**: Never log sensitive field values, even in debug mode

### Review Checklist

When adding fields that handle sensitive data, ensure:
- [ ] Field is marked as sensitive in the schema
- [ ] Documentation clearly indicates the field is sensitive
- [ ] Acceptance tests use appropriate dummy values
- [ ] Field is properly handled in CRUD operations
- [ ] No sensitive data appears in error messages or logs

By following these security practices, we protect users' sensitive information and maintain the trustworthiness of the Scaleway provider.

By following these guidelines, you can ensure your contributions align with the project's standards and maintain consistency across the codebase.
