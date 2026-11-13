The [`scaleway_cockpit_token`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/ephemeral-resources/cockpit_token) Ephemeral Resource creates a temporary Scaleway Cockpit token that is never stored in Terraform state.

Unlike the regular [`scaleway_cockpit_token` Resource](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/cockpit_token), this ephemeral resource creates a token during plan/apply, exposes `secret_key` for use in ephemeral contexts (for example write-only attributes), and deletes the token when Terraform finishes using it (`Close`).

Each Terraform run creates a new token. No token persists between runs when `Close` succeeds.

For more information, see [our guide to using Ephemeral Resources](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-ephemeral-resources), the [Cockpit documentation](https://www.scaleway.com/en/docs/observability/cockpit/), and the [API documentation](https://www.scaleway.com/en/docs/cockpit/api-cli/).
