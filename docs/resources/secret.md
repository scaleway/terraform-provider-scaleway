---
subcategory: "Secrets"
page_title: "Scaleway: scaleway_secret"
---

# Resource: scaleway_secret

The `scaleway_secret` resource allows you to create and manage secrets in Scaleway Secret Manager.

Refer to the Secret Manager [product documentation](https://www.scaleway.com/en/docs/secret-manager/) and [API documentation](https://www.scaleway.com/en/developers/api/secret-manager/) for more information.



## Example Usage

```terraform
# Create a secret

resource "scaleway_secret" "main" {
  name        = "foo"
  description = "barr"
  tags        = ["foo", "terraform"]
}
```

```terraform
# Apply an ephemeral policy on your secret named `foo`.
# In the example below, your secret's lifetime is of 24 hours, your secret versions will expire once they are accessed, and they are disabled after being accessed.

resource "scaleway_secret" "ephemeral" {
  name = "foo"
  ephemeral_policy {
    ttl                   = "24h"
    expires_once_accessed = true
    action                = "disable"
  }
}
```

```terraform
# Create a secret with your own Key Manager encryption key

resource "scaleway_key_manager_key" "main" {
  name        = "my-kms-key"
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  unprotected = true
}

resource "scaleway_secret" "kms" {
  name   = "foo"
  key_id = scaleway_key_manager_key.main.id
}
```




## Argument Reference

The following arguments are supported:

- `name` - (Required) Name of the secret (e.g. `my-secret`).
- `path` - (Optional) Path of the secret, defaults to `/`.
- `protected` - (Optional) True if secret protection is enabled on the secret. A protected secret cannot be deleted, terraform will fail to destroy unless this is set to false.
- `description` - (Optional) Description of the secret (e.g. `my-new-description`).
- `tags` - (Optional) Tags of the secret (e.g. `["tag", "secret"]`).
- `ephemeral_policy` - (Optional) [Ephemeral policy](https://www.scaleway.com/en/docs/identity-and-access-management/secret-manager/concepts/#ephemeral-policy) of the secret. Policy that defines whether/when a secret's versions expire. By default, the policy is applied to all the secret's versions.
    - `ttl` - (Optional) Time frame, from one second and up to one year, during which the secret's versions are valid. Has to be specified in [Go Duration format](https://pkg.go.dev/time#ParseDuration) (ex: "30m", "24h").
    - `expires_once_accessed` - (Optional) True if the secret version expires after a single user access.
    - `action` - (Required) Action to perform when the version of a secret expires. Available values can be found in [SDK constants](https://pkg.go.dev/github.com/scaleway/scaleway-sdk-go@master/api/secret/v1beta1#pkg-constants).
- `type` - (Optional) Type of the secret. If not specified, the type is Opaque. Available values can be found in [SDK Constants](https://pkg.go.dev/github.com/scaleway/scaleway-sdk-go@master/api/secret/v1beta1#pkg-constants).
- `key_id` - (Optional, Forces new resource) ID of the [Key Manager](https://www.scaleway.com/en/docs/key-manager/) key used to encrypt and decrypt the secret's versions. Can be a UUID or a localized ID (`region/UUID`). If not set, Secret Manager uses a Key Manager internal key.
- `region` - (Optional, Computed, Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions)
  in which the resource exists.
- `project_id` - (Optional) The project ID containing is the secret.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `version_count` - The amount of secret versions.
- `status` - The status of the secret.
- `created_at` - Date and time of the secret's creation (in RFC 3339 format).
- `updated_at` - Date and time of the secret's last update (in RFC 3339 format).
- `srn` - The Scaleway Resource Name (SRN) of the secret.

## Import

This section explains how to import a secret using the `{region}/{id}` format.

```bash
terraform import scaleway_secret.main fr-par/11111111-1111-1111-1111-111111111111
```
