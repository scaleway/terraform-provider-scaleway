---
subcategory: "Key Manager"
page_title: "Scaleway: scaleway_key_manager_verify"
---

# Data Source: scaleway_key_manager_verify

The [`scaleway_key_manager_verify`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/ephemeral-resource/key_manager_verify) data source is used to verify a message signature against a message digest with a given key. The key must have its usage set to asymmetric_signing. The message digest must be generated using the same digest algorithm that is defined in the key's algorithm configuration, and encoded as a base64 string.

Refer to the Key Manager [documentation](https://www.scaleway.com/en/docs/key-manager/) and [API documentation](https://www.scaleway.com/en/developers/api/key-manager/) for more information.


## Example Usage

```terraform
# The following commands allow you to:

# - create a key named `my-kms-key`
# - generate a signature for a message using the key created above
# - store the signature in a secret manager secret version
# - verify the signature using the key created above, the digest and the signature retrieved from the secret

// Create a key
resource "scaleway_key_manager_key" "main" {
  name        = "my-kms-key"
  region      = "fr-par"
  usage       = "asymmetric_signing"
  algorithm   = "rsa_pss_2048_sha256"
  unprotected = true
}

// Generate a signature for a message using the key created above
ephemeral "scaleway_key_manager_sign" "main" {
  key_id = scaleway_key_manager_key.main.id
  digest = "base64digest"
  region = "fr-par"
}

resource "scaleway_secret" "main" {
  name = "my-secret"
}

// Store the signature in a secret manager secret version
resource "scaleway_secret_version" "signature" {
  secret_id = scaleway_secret.main.id
  data_wo   = ephemeral.scaleway_key_manager_sign.main.signature
}

data "scaleway_secret_version" "signature" {
  secret_id = scaleway_secret.main.id
  revision  = "1"
}

// Verify the signature using the key created above, the digest and the signature retrieved from the secret
data "scaleway_key_manager_verify" "main" {
  key_id    = scaleway_key_manager_key.main.id
  region    = "fr-par"
  digest    = "base64digest"
  signature = data.scaleway_secret_version.signature.data
}
```




## Argument Reference

- `key_id` -  ID of the key to use for signature verification. Can be a plain UUID or a [regional](../guides/regions_and_zones.md#resource-ids) ID.

- `region` - The [region](../guides/regions_and_zones.md#regions) of the key. If not set, the region is derived from the key_id when possible or from the [provider](../index.md#region) `region` configuration.

- `digest` - Digest of the original signed message. Must be generated using the same algorithm specified in the key’s configuration, and encoded as a base64 string.

- `signature` - The message signature to verify, encoded as a base64 string.

- `valid` - Defines whether the signature is valid. Returns `true` if the signature is valid for the digest and key, and `false` otherwise.

## Attributes Reference

Exported attributes are the ones from `scaleway_key_manager_key` [resource](../resources/key_manager_key.md)
