---
subcategory: "Secrets"
page_title: "Scaleway: scaleway_secret"
---

# scaleway_secret

The `scaleway_secret` data source is used to get information about a specific secret in Scaleway's Secret Manager.

Refer to the Secret Manager [product documentation](https://www.scaleway.com/en/docs/identity-and-access-management/secret-manager/) and [API documentation](https://www.scaleway.com/en/developers/api/secret-manager/) for more information.

## Crrate a secret and get its information

The following commands show you how to:

- create a secret named `foo` with the description `barr`
- retrieve the secret information using the secret's ID
- retrieve the secret information using the secret's name

```hcl
// Create a secret
resource "scaleway_secret" "main" {
  name        = "foo"
  description = "barr"
}

// Get the secret information specified by the secret ID
data "scaleway_secret" "my_secret" {
  secret_id = "11111111-1111-1111-1111-111111111111"
}

// Get the secret information specified by the secret name
data "scaleway_secret" "by_name" {
  name = "your_secret_name"
}
```

## Arguments reference

This section lists the arguments that can be provided to the `scaleway_secret` data source:


- `name` - (Optional) The name of the secret.
  Only one of `name` and `secret_id` should be specified.

- `path` - (Optional) The path of the secret.
  Conflicts with `secret_id`.

- `secret_id` - (Optional) The ID of the secret.
  Only one of `name` and `secret_id` should be specified.

- `organization_id` - (Optional) The ID of the Scaleway Organization the Project is associated with. If no default `organization_id` is set, it must be set explicitly in this data source.

- `region` - (Defaults to the region specified in the [provider's configuration](../index.md#region)). The [region](../guides/regions_and_zones.md#regions) in which the secret exists.

- `project_id` - (Optional. Defaults to the Project specified in [provider's configuration](../index.md#project_id)). The ID of the
  Project the secret is associated with.


## Attributes reference

Exported attributes are the ones from the `scaleway_secret` [resource](../resources/secret.md).
