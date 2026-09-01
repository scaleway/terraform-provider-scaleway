The [`scaleway_iam_api_key`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/ephemeral-resource/iam_api_key) Ephemeral Resource is used to create and manage Scaleway API Keys. An API key can be associated with either an application or a user.

~> **Important:** This ephemeral resource is currently experimental and may evolve as we refine its functionality. Unlike the regular [`scaleway_iam_api_key` Resource](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/iam_api_key), this ephemeral resource is not stored in Terraform state and is not managed by Terraform after creation.

## Lifecycle Management

This ephemeral resource supports an `ephemeral_lifecycle` attribute that controls how the API key is managed:

- `persist` (default): The API key and its annotations are not deleted when the ephemeral resource is closed. The API key persists after Terraform operations complete. This is the default Terraform behaviour for ephemeral resources (currently in Terraform v1.18). If an **identifier** is set, on subsequent applies the created API key will be retrieved and will not be recreated.
- `delete`: The API key and its annotations are deleted when the ephemeral resource is closed. Use this for temporary credentials that should be cleaned up automatically, or to clean up an ephemeral that has an **identifier** before removing it from your configuration.
- `replace`: Any existing API key with the same **identifier** is deleted and a new one is created. Use this to ensure a fresh API key on each run. Requires one of `annotation_identifier` or `description_identifier` to be set.

## Identifier Strategies

To enable resource lookup and management, this ephemeral resource supports two identifier strategies:

- **Annotation-based** (`annotation_identifier`): Resources are identified by a unique annotation value. This is the recommended approach for most use cases.
- **Description-based** (`description_identifier`): Resources are identified by a unique description string.

~> **Note:** When using `annotation_identifier` or `description_identifier`, ensure the identifier value is unique across your organization for this resource type to avoid conflicts.

For more information, see [our guide to using Ephemeral Resources](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-ephemeral-resources), the [IAM documentation](https://www.scaleway.com/en/docs/iam/), and the [API documentation](https://www.scaleway.com/en/developers/api/iam/).
