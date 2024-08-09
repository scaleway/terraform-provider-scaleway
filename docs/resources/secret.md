---
subcategory: "Secrets"
page_title: "Scaleway: scaleway_secret"
---

# Resource: scaleway_secret

Creates and manages Scaleway Secrets.
For more information, see [the documentation](https://www.scaleway.com/en/developers/api/secret-manager/).

## Example Usage

### Basic

```terraform
resource "scaleway_secret" "main" {
  name        = "foo"
  description = "barr"
  tags        = ["foo", "terraform"]
}
```

### Ephemeral Policy

```terraform
resource "scaleway_secret" "ephemeral" {
  name = "foo"
  ephemeral_policy {
    ttl = "24h"
    expires_once_accessed = true
    action = "disable"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the secret (e.g. `my-secret`).
- `path` - (Optional) Path of the secret, defaults to `/`.
- `protected` - (Optional) True if secret protection is enabled on the secret. A protected secret cannot be deleted, terraform will fail to destroy unless this is set to false.
- `description` - (Optional) Description of the secret (e.g. `my-new-description`).
- `tags` - (Optional) Tags of the secret (e.g. `["tag", "secret"]`).
- `ephemeral_policy` - (Optional) Ephemeral policy of the secret. Policy that defines whether/when a secret's versions expire. By default, the policy is applied to all the secret's versions.
    - `ttl` - (Optional) Time frame, from one second and up to one year, during which the secret's versions are valid. Has to be specified in [Go Duration format](https://pkg.go.dev/time#ParseDuration) (ex: "30m", "24h").
    - `expires_once_accessed` - (Optional) True if the secret version expires after a single user access.
    - `action` - (Required) Action to perform when the version of a secret expires. Available values can be found in [SDK constants](https://pkg.go.dev/github.com/scaleway/scaleway-sdk-go@master/api/secret/v1beta1#pkg-constants).
- `type` - (Optional) Type of the secret. If not specified, the type is Opaque. Available values can be found in [SDK Constants](https://pkg.go.dev/github.com/scaleway/scaleway-sdk-go@master/api/secret/v1beta1#pkg-constants).
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions)
  in which the resource exists.
- `project_id` - (Optional) The project ID containing is the secret.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `version_count` - The number of versions for this Secret.
- `status` - The status of the Secret.
- `created_at` - Date and time of secret's creation (RFC 3339 format).
- `updated_at` - Date and time of secret's last update (RFC 3339 format).

## Import

The Secret can be imported using the `{region}/{id}`, e.g.

```bash
terraform import scaleway_secret.main fr-par/11111111-1111-1111-1111-111111111111
```
