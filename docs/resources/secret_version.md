---
page_title: "Scaleway: scaleway_secret_version"
description: |-
Manages Scaleway Secret Versions
---

# scaleway_secret

Creates and manages Scaleway Secret Versions.
For more information, see [the documentation](https://developers.scaleway.com/en/products/secret_manager/api/v1alpha1/#secret-versions-079501).

## Examples

### Basic

```hcl
resource "scaleway_secret" "main" {
  name        = "foo"
  description = "barr"
  tags        = ["foo", "terraform"]
}

resource "scaleway_secret_version" "v1" {
  description = "version1"
  secret_id   = scaleway_secret.main.id
  data        = "my_new_secret"
}
```

## Arguments Reference

The following arguments are supported:

- `secret_id` - (Required) The Secret ID associated wit the secret version.
- `data` - (Optional) The data payload of the secret version encode on base64(e.g. `my-secret-version-payload`).
~> **Important:** Updates to `data` will force new the secret version. Be aware that this is a sensitive attribute. For more information, see [Sensitive Data in State](https://developer.hashicorp.com/terraform/language/state/sensitive-data).
- `description` - (Optional) Description of the secret version (e.g. `my-new-description`).
- `with_access` - (Optional) Enable access to the secret.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions)
  in which the resource exists.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `revision` - The revision for this Secret Version.
- `status` - The status of the Secret Version.
- `created_at` - Date and time of secret version's creation (RFC 3339 format).
- `updated_at` - Date and time of secret version's last update (RFC 3339 format).

## Import

The Secret Version can be imported using the `{region}/{id}/{revision}`, e.g.

~> **Important:** Be aware if you import with revision `latest` you will overwrite your version if you used before.

```bash
$ terraform import scaleway_secret.main fr-par/11111111-1111-1111-1111-111111111111/2
```