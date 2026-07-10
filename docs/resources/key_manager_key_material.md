---
subcategory: "Key Manager"
page_title: "Scaleway: scaleway_key_manager_key_material"
---

# Resource: scaleway_key_manager_key_material
Import externally generated key material into Key Manager to derive a new cryptographic key. The key's origin must be external.

-> **Security Best Practice:**
For enhanced security, we recommend using the [`key_material_wo`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) and [`salt_wo`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) write-only arguments instead of the regular `key_material` and `salt` arguments. This ensures your sensitive cryptographic material is never stored in Terraform state files, providing superior protection against accidental exposure. Write-Only arguments are supported in Terraform 1.11.0 and later.

-> **Note:** When using write-only arguments (`key_material_wo` and `salt_wo`), you must also provide the corresponding version fields (`key_material_wo_version` and `salt_wo_version`) to enable proper resource lifecycle management.



## Example Usage

```terraform
resource "scaleway_key_manager_key" "main" {
  name        = "my-external-key"
  description = "Key with externally imported material"
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  origin      = "external"
  region      = "fr-par"
}

resource "random_bytes" "key_material" {
  length = 32 # 256-bit key for AES-256
}

resource "scaleway_key_manager_key_material" "main" {
  key_id                  = scaleway_key_manager_key.main.id
  key_material_wo         = base64encode(random_bytes.key_material.base64)
  key_material_wo_version = 1
}
```

```terraform
resource "scaleway_key_manager_key" "main" {
  name        = "my-external-key"
  description = "Key with externally imported material and salt"
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  origin      = "external"
  region      = "fr-par"
}

resource "random_bytes" "key_material" {
  length = 32 # 256-bit key for AES-256
}

resource "random_bytes" "salt" {
  length = 16 # 128-bit salt
}

resource "scaleway_key_manager_key_material" "main" {
  key_id                  = scaleway_key_manager_key.main.id
  key_material_wo         = base64encode(random_bytes.key_material.base64)
  key_material_wo_version = 1
  salt_wo                 = base64encode(random_bytes.salt.base64)
  salt_wo_version         = 1
}
```



## Argument Reference

The following arguments are supported:

- `key_id` - (Required, ForceNew) The ID of the key to import key material into. The key's origin must be external (UUID format). Can be a plain UUID or a regional ID.
- `key_material` - (Optional, Sensitive, ForceNew) The key material to import. The key material is a random sequence of bytes used to derive a cryptographic key. Can be provided as raw bytes or a base64-encoded string (the provider will automatically normalize the input).
- `key_material_wo` - (Optional, Write-only) The key material to import in write-only mode. The key material is a random sequence of bytes used to derive a cryptographic key. Must be provided as a base64-encoded string. The key material will not be stored in the Terraform state. Either `key_material` or `key_material_wo` must be specified.
- `key_material_wo_version` - (Optional, ForceNew) Version number to track changes to the write-only key material. Increment this value to recreate the resource with new key material. Required when using `key_material_wo`.
- `salt` - (Optional, Sensitive, ForceNew) Optional salt for key derivation. A salt is random data added to key material to ensure unique derived keys, even if the input is similar. It helps strengthen security when the key material has low randomness (low entropy). Can be provided as raw bytes or a base64-encoded string (the provider will automatically normalize the input). Only one of `salt` or `salt_wo` can be specified.
- `salt_wo` - (Optional, Write-only) Optional salt for key derivation in write-only mode. A salt is random data added to key material to ensure unique derived keys. Must be provided as a base64-encoded string. The salt will not be stored in the Terraform state. Only one of `salt` or `salt_wo` can be specified.
- `salt_wo_version` - (Optional, ForceNew) Version number to track changes to the write-only salt. Increment this value to recreate the resource with new salt. Required when using `salt_wo`.
- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions) of the key. If not set, the region is derived from the key_id when possible or from the provider configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the key material resource (same as key_id).
- `key_state` - The current state of the key (enabled, disabled, pending_key_material).
- `origin` - The origin of the key (should be 'external').
