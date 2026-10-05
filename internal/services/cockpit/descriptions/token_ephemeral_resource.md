The [`scaleway_cockpit_token`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/ephemeral-resource/cockpit_token) Ephemeral Resource is used to create temporary Scaleway Cockpit tokens. Unlike the regular [`scaleway_cockpit_token` Resource](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/cockpit_token), this ephemeral resource is not stored in Terraform state and is deleted at the end of each Terraform run (plan or apply).

Each `terraform apply` will create a new token. The token is automatically deleted when Terraform finishes using it. This ensures no tokens persist between runs.

For more information, see [our guide to using Ephemeral Resources](https://developer.hashicorp.com/terraform/language/resources/ephemeral), the [Cockpit documentation](https://www.scaleway.com/en/docs/observability/cockpit/), and the [API documentation](https://www.scaleway.com/en/developers/api/cockpit/).
