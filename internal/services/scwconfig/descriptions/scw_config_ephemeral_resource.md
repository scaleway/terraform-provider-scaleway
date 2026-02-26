The [`scaleway_config`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/ephemeral-resource/config) Ephemeral Resource is used to retrieve Scaleway configuration information including project ID, organization ID, access key, secret key, zone, and region.

Contrary to the scw_config Data Source, this is an Ephemeral Resource: it is not stored in the Terraform state file. The configuration information is retrieved fresh with each Terraform apply.

Refer to the Scaleway [documentation](https://www.scaleway.com/en/docs/) and [API documentation](https://www.scaleway.com/en/developers/api/) for more information.
