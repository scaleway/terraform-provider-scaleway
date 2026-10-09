# Terraform Provider for Scaleway

- [Provider Documentation Website](https://www.terraform.io/docs/providers/scaleway/index.html)
- Slack: [Scaleway-community Slack][slack-scaleway] ([#terraform][slack-terraform])

[slack-scaleway]: https://slack.scaleway.com/
[slack-terraform]: https://scaleway-community.slack.com/app_redirect?channel=terraform

## Requirements

All the development tools (Go, Terraform, OpenTofu, golangci-lint, ...) and their versions are pinned in [`mise.toml`](mise.toml) and installed with [mise](https://mise.jdx.dev/):

```sh
mise install
```

This single configuration drives the CI as well, so you can rely on it being complete.

## Getting started

### On your machine (recommended)

Install [mise](https://mise.jdx.dev/getting-started.html), then clone the repository and install the pinned tools:

```sh
git clone git@github.com:scaleway/terraform-provider-scaleway.git
cd terraform-provider-scaleway
mise install
```

### As an alternative: devcontainer

If you prefer an isolated environment, a [devcontainer](.devcontainer/devcontainer.json) with `mise` and the VS Code setup already configured is provided as a convenience:

- Open the repository in [VS Code](https://code.visualstudio.com/docs/devcontainers/tutorial) and choose **Reopen in Container** (requires the *Dev Containers* extension).
- The tools declared in `mise.toml` are resolved lazily the first time you run a task; run `mise install` once in the container to prefetch them all.

All the `mise run ...` commands below work the same in both setups.

## Building The Provider

```sh
mise run build:provider
```

The binary is written to `./terraform-provider-scaleway` (override the location with `mise run build:provider --output <path>`).

## Using the provider

See the [Scaleway Provider Documentation](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs) to get started using the Scaleway provider.

## Developing the Provider

Run `mise tasks` to list every available task (`build:*`, `test:*`, `lint:*`, `gen:*`, ...). The most common ones are:

```sh
mise run build:provider       # build the provider binary
mise run test:provider        # run the unit tests
mise run test:provider --acceptance              # run the acceptance tests against the recorded mocks
mise run test:provider --acceptance --cassettes  # run the acceptance tests against the real API and record cassettes
mise run lint:go              # lint the Go code
mise run gen:doc              # regenerate the provider documentation
mise run default              # fast local loop: build + test + lint the provider
```

If you want to use your locally built provider with Terraform instead of the published one, you can take advantage of [development overrides](https://www.terraform.io/cli/config/config-file#development-overrides-for-provider-developers) (`dev_overrides`) in the Terraform CLI configuration, and point Terraform/OpenTofu at that configuration with `TF_CLI_CONFIG_FILE` set in the git-ignored `.env` file.

Please refer to the [TESTING.md](TESTING.md#using-a-locally-built-provider) for testing.
