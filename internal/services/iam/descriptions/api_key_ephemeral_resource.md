The [`scaleway_iam_api_key`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/ephemeral-resource/iam_api_key) Ephemeral Resource is used to create and manage Scaleway API Keys. An API key can be associated with either an application or a user.

Contrary to the iam_api_key Resource, this is an Ephemeral Resource: it is not stored in the Terraform state file, and is therefore not managed by Terraform and needs to be managed separately. A new Scaleway API key will be created with each Terraform apply. You may set the `expires_at` attribute for the key to be automatically deleted at a set date.

Refer to the IAM [documentation](https://www.scaleway.com/en/docs/iam/) and [API documentation](https://www.scaleway.com/en/developers/api/iam/) for more information.