# Testing the module

## Prerequisites

The development tools are pinned in [`mise.toml`](mise.toml) and installed with [mise](https://mise.jdx.dev/):

```sh
mise install
```

## Unit testing

You can test the provider, by running `mise run test:provider`.

```sh
mise run test:provider
```

## Acceptance testing

Acceptance test are made to test the terraform module with real API calls so they will create real resources that will be invoiced.
But in order to run faster tests and avoid bad surprise at the end of the month we shipped the project with mocks (recorded with [go-vcr](https://github.com/dnaeon/go-vcr)).

### Running the acceptance tests with mocks

By default, mocks are used during acceptance tests.

```sh
mise run test:provider --acceptance
```

### Running the acceptance tests on real resources

:warning: This will cost money.

```sh
mise run test:provider --acceptance --cassettes
```

`--cassettes` records new cassettes by making real API calls.

### Credentials

The provider resolves credentials the same way the `scw` CLI does, from `~/.config/scw/config.yaml` (the file created by `scw init`), environment variables taking precedence.
The recommended setup is therefore to configure your Scaleway account with `scw init`.
You can select a specific profile with `SCW_PROFILE=...` or point at an alternate config file with `SCW_CONFIG_PATH=...`.

Environment variables are also supported:

```sh
export SCW_ACCESS_KEY=SCWXXXXXXXXXXXXXXXXX
export SCW_SECRET_KEY=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
export SCW_DEFAULT_PROJECT_ID=XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
```

For testing the domain API, it will use the first available domain in your domains list. You need to have a valid domain.

You can force the test domain with an environment var:

```sh
export TF_TEST_DOMAIN=your-domain.tld
```

For testing a domain zone you can force the following environment var:

```sh
export TF_TEST_DOMAIN_ZONE=your-zone
```

To ease debugging you can also set `TF_LOG=DEBUG` or append `--debug` to the test task (sets `SCW_DEBUG=true`).

Running a single test:

```sh
mise run test:provider --acceptance --run 'TestAccScalewayDataSourceRDBInstance_Basic' --timeout 120m
```

To record the cassette of a single test against the real API:

```sh
mise run test:provider --acceptance --cassettes --run 'TestAccScalewayDataSourceRDBInstance_Basic' --timeout 120m
```

## Compressing the cassettes

We record interactions with the Scaleway API in cassettes, which are stored in the `testdata` directory of each service.
Each wait function used in the resources will perform several requests to the API for pulling a resource state, which can lead to large cassettes.
We use a compressor to reduce the size of these cassettes once they are recorded.
By doing so, tests can run faster and the cassettes are easier to read.

To use the compressor on a given cassette, run the following command:

```sh
go run -v ./cmd/vcr-compressor internal/services/rdb/testdata/acl-basic.cassette
```

## Using a locally built provider

`mise run build:provider` builds the provider binary at `./terraform-provider-scaleway` (override the path with `-o <path>`, see `mise run build:provider --help`).

To use this local build with `terraform` instead of the version published on the registry, use [development overrides](https://developer.hashicorp.com/terraform/cli/config/config-file#development-overrides-for-provider-developers) (`dev_overrides`) in the Terraform CLI configuration:

```hcl
# .terraformrc
provider_installation {
  dev_overrides {
    "scaleway/scaleway" = "/absolute/path/to/terraform-provider-scaleway"
  }
  direct {}
}
```

The directory must contain the built binary named `terraform-provider-scaleway` (the default output of `build:provider`). The `direct {}` block keeps every other provider installing from its registry as usual — without it, Terraform refuses to download any provider at all. The key can be written as `scaleway/scaleway` or the full `registry.terraform.io/scaleway/scaleway`.

With `dev_overrides` in effect, Terraform prints a warning on every `plan`/`apply` that the local binary may not match any released version, and `version` constraints in your configuration are ignored. This only affects manual CLI runs: the acceptance test harness launches the provider in-process and is unaffected.

### Selecting the CLI configuration file with `.env`

Keep the override out of your global `~/.terraformrc` so it only applies to this repository. mise loads the git-ignored `.env` file at the repository root (see `[env] _.file = '.env'` in `mise.toml`) and exports it into every task and every tool launched through mise here, so point Terraform/OpenTofu at a dedicated CLI configuration file from there:

```
# .env (git-ignored, loaded by mise)
TF_CLI_CONFIG_FILE=/absolute/path/to/terraform-provider-scaleway/.terraformrc
```

`TF_CLI_CONFIG_FILE` is the variable both tools read: Terraform defaults to `~/.terraformrc`, OpenTofu to `~/.tofurc`. Only the path matters; the file name and extension are irrelevant. (OpenTofu does not read its `TOFU_CLI_CONFIG_FILE` variant at the pinned 1.12.)

Once set, `terraform`/`tofu` invoked through mise in this directory (or any `mise run ...` task) resolves the scaleway provider from your local build instead of the registry:

```sh
mise run build:provider                      # binary written to ./terraform-provider-scaleway
terraform -chdir=/path/to/your/config init
terraform -chdir=/path/to/your/config plan   # prints: Provider development overrides are in effect
```

Note that `tofu init` still installs the provider from the registry even with the override in effect, and prints `Skip tofu init when using provider development overrides`. That copy is unused at `plan`/`apply` time, so it is harmless — the override only shows up from `plan` onward.
