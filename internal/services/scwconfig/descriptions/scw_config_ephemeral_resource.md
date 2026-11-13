The [`scaleway_config`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/ephemeral-resource/config) Ephemeral Resource is used to retrieve Scaleway configuration information including project ID, organization ID, access key, secret key, zone, and region.

Contrary to the scw_config Data Source, this is an Ephemeral Resource: it is not stored in the Terraform state file. The configuration information is retrieved fresh with each Terraform apply.

For more information, see [our guide to using Ephemeral Resources with Terraform Scaleway Provider](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-ephemeral-resources), the Scaleway [documentation](https://www.scaleway.com/en/docs/), and the [API documentation](https://www.scaleway.com/en/developers/api/).
