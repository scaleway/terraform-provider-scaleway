---
subcategory: "Key Manager"
page_title: "Scaleway: scaleway_key_manager_key"
---

# scaleway_key_manager_key

Gets information about a Key Manager Key. For more information, refer to the [Key Manager API documentation](https://www.scaleway.com/en/developers/api/key-manager/#path-keys-get-key-metadata).

## Example Usage

### Create a key and get its information

The following commands allow you to:

- create a key named `my-kms-key`
- retrieve the key's information using the key's ID

```hcl
// Create a key
resource "scaleway_key_manager_key" "symmetric" {
  name        = "my-kms-key"
  region      = "fr-par"
  project_id  = "your-project-id" # optional, will use provider default if omitted
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  description = "Key for encrypting secrets"
  tags        = ["env:prod", "kms"]
  unprotected = true

  rotation_policy {
    rotation_period = "720h" # 30 days
  }
}

// Get the key information by its ID
data "scaleway_key_manager_key" "byID" {
  key_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `key_id` -  ID of the key to target. Can be a plain UUID or a [regional](../guides/regions_and_zones.md#resource-ids) ID.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the key was created.

## Attributes Reference

Exported attributes are the ones from `scaleway_key_manager_key` [resource](../resources/key_manager_key.md)
