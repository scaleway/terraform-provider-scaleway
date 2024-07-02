---
subcategory: "Secrets"
page_title: "Scaleway: scaleway_secret"
---

# Resource: scaleway_secret

The `scaleway_secret` resource allows you to create and manage secrets in Scaleway Secret Manager.

Refer to the Secret Manager [product documentation](https://www.scaleway.com/en/docs/identity-and-access-management/secret-manager/) and [API documentation](https://www.scaleway.com/en/developers/api/secret-manager/) for more information.

## Create a secret

The following command allows you to create a secret named `foo` with a description (`barr`), and tags (`foo` and `terraform`).

```terraform
resource "scaleway_secret" "main" {
  name        = "foo"
  description = "barr"
  tags        = ["foo", "terraform"]
}
```

## Arguments reference

The following arguments are supported:

- `name` - (Required) Name of the secret (e.g. `my-secret`).
- `path` - (Optional) Path of the secret, defaults to `/`.
- `description` - (Optional) A description of the secret (e.g. `my-new-description`).
- `tags` - (Optional) Tags of the secret (e.g. `["tag", "secret"]`).
- `region` - (Defaults to the region specififed in the [provider configuration](../index.md#region)) The [region](../guides/regions_and_zones.md#regions) where the resource exists.
- `project_id` - (Optional) The ID of the Project containing the secret.

## Attributes reference

This section lists the attributes that are exported when the `scaleway_secret` resource is created:

- `version_count` - The amount of secret versions.
- `status` - The status of the secret.
- `created_at` - Date and time of the secret's creation (in RFC 3339 format).
- `updated_at` - Date and time of the secret's last update (in RFC 3339 format).

## Import

This section explains how to import a secret using the `{region}/{id}` format.

```bash
$ terraform import scaleway_secret.main fr-par/11111111-1111-1111-1111-111111111111
```
