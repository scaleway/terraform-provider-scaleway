---
page_title: "Using Ephemeral Resources Guide"
---
# Using Ephemeral Resources with the Terraform Scaleway Provider

Ephemeral resources in Terraform allow you to access sensitive data during Terraform operations without storing that data in the Terraform state file. This ensures your sensitive credentials are never stored in Terraform state files, providing superior protection against accidental exposure. This guide explains how to use ephemeral resources in the Scaleway Terraform Provider.

For more information, see the [official HashiCorp documentation for Ephemeral Resources](https://developer.hashicorp.com/terraform/plugin/framework/resources/ephemeral).

## What are Ephemeral Resources?

Ephemeral resources are special Terraform resources that are used during Terraform operations but are **not stored** in the Terraform state file. They are designed to temporarily access and read sensitive data that should not persist in state, such as secret values, passwords, or temporary credentials. Similarly to data sources, ephemeral resources are queried during each Terraform operation. To achieve maximum security and prevent any sensitive data from being stored in state, ephemeral resources should be used in conjunction with [write-only arguments](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments).

## Resources Supporting Ephemeral Resources

The Scaleway Terraform Provider supports ephemeral resources for several services:

### Secret Manager Resources
- [**`scaleway_secret_version`**](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/ephemeral-resources/secret_version)

### Key Manager Resources
- [**`scaleway_key_manager_sign`**](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/ephemeral-resources/key_manager_sign)
- [**`scaleway_key_manager_encrypt`**](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/ephemeral-resources/key_manager_encrypt)
- [**`scaleway_key_manager_decrypt`**](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/ephemeral-resources/key_manager_decrypt)
- [**`scaleway_key_manager_generate_data_key`**](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/ephemeral-resources/key_manager_generate_data_key)

## How to use Ephemeral Resources in Scaleway Provider

The Scaleway Terraform Provider implements ephemeral resources using the `ephemeral` block type. These resources are used to temporarily access sensitive data during Terraform operations.

Ephemeral resources are typically used to access data that was stored using write-only arguments or other sensitive data sources. Ephemeral resource attributes can only be referenced in other ephemeral resources or in write-only arguments.

### Example: Accessing Secret Version Data

```terraform
# Create a secret
resource "scaleway_secret" "main" {
  name        = "my-secret"
  description = "my-secret-description"
}

# Create a secret version using write-only argument
resource "scaleway_secret_version" "v1" {
  description     = "A secret version with write-only data"
  secret_id       = scaleway_secret.main.id
  data_wo         = "my_secret_data" # Not stored in state
  data_wo_version = 1
}

# Access the secret version data using an ephemeral resource (not stored in the state)
ephemeral "scaleway_secret_version" "data_v1" {
  secret_id  = scaleway_secret.main.id
  revision   = "1"
  depends_on = [scaleway_secret_version.v1]
}

# Use the ephemeral data in other resources
resource "scaleway_mongodb_instance" "example" {
  name                = "my-mongodb-instance"
  version             = "7.0.12"
  node_type           = "MGDB-PLAY2-NANO"
  node_number         = 1
  user_name           = "my_initial_user"
  password_wo         = ephemeral.scaleway_secret_version.data_v1.data
  password_wo_version = 1
}
```

### Example: Encrypting a plaintext with key_manager_encrypt ephemeral resource
```terraform
# Create an encryption key that will be used to encrypt a plaintext
resource "scaleway_key_manager_key" "main" {
  name        = "my-encryption-key"
  region      = "fr-par"
  usage       = "symmetric_encryption"
  algorithm   = "aes_256_gcm"
  unprotected = true
}

# Encrypt the plaintext with the created encryption key using the key_manager_encrypt Ephemeral resource (not stored in state)
ephemeral "scaleway_key_manager_encrypt" "main" {
  key_id     = scaleway_key_manager_key.main.id
  plaintext  = "This is a sensitive plaintext"
  region     = "fr-par"
}

# The encrypted ciphertext can be stored in a scaleway_secret, using a write-only argument
# Create a secret that will be used to store the encrypted ciphertext
resource "scaleway_secret" "main" {
  name        = "my-secret"
}

# Store the ciphertext in a secret_version using the data write-only argument
resource "scaleway_secret_version" "v1" {
  description     = "A secret version containing an encrypted ciphertext"
  secret_id       = scaleway_secret.main.id
  data_wo         = ephemeral.scaleway_key_manager_encrypt.main.ciphertext # Not stored in state
  data_wo_version = 1
}
```
