---
page_title: "Using Write-Only Arguments Guide"
---
# Using Write-Only Arguments with the Terraform Scaleway Provider

Write-only arguments in Terraform allow you to handle sensitive data that should not be stored in the Terraform state file. This ensures your sensitive credentials are never stored in Terraform state files, providing superior protection against accidental exposure. This guide explains how to use write-only arguments in the Scaleway Terraform Provider.
Write-Only arguments are supported in Terraform 1.11.0 and later.

For more information, see the [official Terraform documentation for Write-only Arguments](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only).

## What are Write-Only Arguments?

Write-only arguments are special attributes that are used during resource creation and updates, but are **not stored** in the Terraform state, and **not displayed** in Terraform plans or outputs. They are ideal for sensitive data like passwords, API keys, or secret values.

## Resources Supporting Write-Only Arguments

The Scaleway Terraform Provider supports write-only arguments in several resources:

### IAM Resources

- [**`scaleway_iam_user`**: `password_wo`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/iam_user#password_wo-1)

### Secret Manager Resources

- [**`scaleway_secret_version`**: `data_wo`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/secret_version#data_wo-1)

### Database Resources

- [**`scaleway_rdb_instance`**: `password_wo`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/rdb_instance#password_wo-2)
- [**`scaleway_rdb_user`**: `password_wo`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/rdb_user#password_wo-3)
- [**`scaleway_redis_cluster`**: `password_wo`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/redis_cluster#password_wo-4)
- [**`scaleway_mongodb_instance`**: `password_wo`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/mongodb_instance#password_wo-5)
- [**`scaleway_mongodb_user`**: `password_wo`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/mongodb_user#password_wo-6)

### Inference Resources

- [**`scaleway_inference_model`**: `secret_wo`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/inference_model#secret_wo-1)

### Baremetal Resources

- [**`scaleway_baremetal_server`**: `password_wo`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/baremetal_server#password_wo-5) and [`service_password_wo`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/baremetal_server#service_password_wo-1)

## How to use Write-Only Arguments in Scaleway Provider

The Scaleway Terraform Provider implements write-only arguments using the following pattern:

1. **Write-Only Attribute**: The sensitive value (e.g., `password_wo`, `data_wo`)
2. **Version Attribute**: A companion version number (e.g., `password_wo_version`, `data_wo_version`)

## Creating a resource with a Write-Only argument

When creating a resource with a write-only argument, you need to provide both the write-only attribute and its version:

### Example: Creating a Secret Version with Write-Only Data

```terraform
resource "scaleway_secret_version" "sensitive_data" {
  secret_id       = scaleway_secret.main.id
  data_wo         = "my-super-secret-value"  # This will NOT be stored in state
  data_wo_version = 1
  description     = "Sensitive secret using write-only mode"
}
```

## Updating a resource's Write-Only argument

Since write-only arguments are not stored in the Terraform state, they cannot be updated by simply changing the value in the configuration because the attribute gets nulled out.

To update a write-only attribute, you must:

1. **Change the write-only value** (e.g., `password_wo`, `data_wo`)
2. **Increment the version number** (e.g., `password_wo_version`, `data_wo_version`)

### Example: Updating Write-Only Secret Version Data

```terraform
resource "scaleway_secret_version" "sensitive_data" {
  secret_id      = scaleway_secret.main.id
  data_wo        = "my-new-super-secret-value"  # Updated secret
  data_wo_version = 2  # Version incremented from 1 to 2
  description    = "Updated sensitive secret"
}
```

## Retrieving a value set with a Write-Only argument

Write-only attributes cannot be read or referenced directly in Terraform.
To work with these values, you must use a secure storage system like [Scaleway Secret Manager](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/secret) that allows both writing and reading sensitive data.
To prevent sensitive data from being stored in state, use the [scaleway_secret_version resource](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/secret_version) with the [data_wo argument](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/secret_version#data_wo-1)to write the secret, then use the [scaleway_secret_version ephemeral resource](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/ephemeral-resources/secret_version) to read it back when needed.

For more information about Ephemeral Resources, see our [guide to using Ephemeral Resources with Terraform Scaleway Provider](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-ephemeral-resources).

### Example: Using an ephemeral resource to retrieve a value set with a write-only argument

```terraform
# Create a secret
resource "scaleway_secret" "main" {
  name        = "my-secret"
  description = "my-secret-description"
}

# Create a secret version, using the write-only data_wo
resource "scaleway_secret_version" "v1" {
  description     = "version1"
  secret_id       = scaleway_secret.main.id
  data_wo         = "my_super_secret_data" # Not stored in the state
  data_wo_version = 1
}

# Access the secret version revision using the ephemeral resource (not stored in state)
ephemeral "scaleway_secret_version" "data_v1" {
  secret_id  = scaleway_secret.main.id
  revision   = "1"
  depends_on = [scaleway_secret_version.v1]
}

# Create a resource (e.g., a MongoDB instance) using its write-only argument to pass the sensitive data
resource "scaleway_mongodb_instance" "password_wo_instance" {
  name                = "my-mongodb-instance"
  version             = "7.0.12"
  node_type           = "MGDB-PLAY2-NANO"
  node_number         = 1
  user_name           = "my_initial_user"
  password_wo         = ephemeral.scaleway_secret_version.data_v1.data
  password_wo_version = 1
}
```

## Write-Only vs Regular Sensitive Attributes

The Scaleway Provider offers both approaches for handling sensitive data.

### Regular Sensitive Attributes (Stored in State)

Regular attributes' values persist in the state and can be referenced by other resources. They are easier to manage, but the sensitive data is stored and visible in the Terraform state. State files must be carefully secured.

```terraform
resource "scaleway_secret_version" "sensitive_data" {
  secret_id     = scaleway_secret.main.id
  data          = "MyNonCriticalS3cr3tP@ssw0rd!"  # Stored in state
  description   = "Sensitive secret using the regular data argument"
}
```

### Write-Only Attributes (Not Stored in State)

Write-only attributes' values are never stored in Terraform state nor visible in plans or outputs. They are therefore more secure for highly sensitive information, but their values cannot be referenced by other resources.

```terraform
resource "scaleway_secret_version" "sensitive_data" {
  secret_id         = scaleway_secret.main.id
  data_wo           = "MyS3cr3tP@ssw0rd!"  # NOT stored in state
  data_wo_version   = 1
  description       = "Sensitive secret using the write-only argument"
}
```
