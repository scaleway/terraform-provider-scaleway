---
subcategory: "Secrets"
page_title: "Scaleway: scaleway_secret_version"
---

# scaleway_secret_version

The `scaleway_secret_version` data source is used to get information about a specific version of a secret stored in Scaleway's Secret Manager.

Refer to the Secret Manager [product documentation](https://www.scaleway.com/en/docs/identity-and-access-management/secret-manager/) and [API documentation](https://www.scaleway.com/en/developers/api/secret-manager/) for more information.


## Use Secret Manager

The following commands show you how to:

- create a secret named `fooii`
- create a new version of `fooii` containing data (`your_secret`)
- retrieve the secret version specified by the secret ID and the desired version
- retrieve the secret version specified by the secret name and the desired version

The output blocks display the sensitive data contained in your secret version.


```hcl
# Create a secret named fooii
resource "scaleway_secret" "main" {
  name        = "fooii"
  description = "barr"
}

# Create a version of fooii containing data
resource "scaleway_secret_version" "main" {
  description = "your description"
  secret_id   = scaleway_secret.main.id
  data        = "your_secret"
}

# Retrieve the secret version specified by the secret ID and the desired version
data "scaleway_secret_version" "data_by_secret_id" {
  secret_id  = scaleway_secret.main.id
  revision   = "1"
  depends_on = [scaleway_secret_version.main]
}

# Retrieve the secret version specified by the secret name and the desired version
data "scaleway_secret_version" "data_by_secret_name" {
  secret_name = scaleway_secret.main.name
  revision    = "1"
  depends_on  = [scaleway_secret_version.main]
}

# Display sensitive data
output "scaleway_secret_access_payload" {
  value = data.scaleway_secret_version.data_by_secret_name.data
}

# Display sensitive data
output "scaleway_secret_access_payload_by_id" {
  value = data.scaleway_secret_version.data_by_secret_id.data
}
```

## Arguments reference

This section lists the arguments that can be provided to the scaleway_secret_version data source:

- `secret_id` - (Optional) The ID of the secret associated with the secret version. Only one of `secret_id` and `secret_name` should be specified.

- `secret_name` - (Optional) The name of the secret associated with the secret version.
  Only one of `secret_id` and `secret_name` should be specified.

- `revision` - The revision for this secret version. Refer to alternative values (ex: `latest`) in the [API documentation](https://www.scaleway.com/en/developers/api/secret-manager/#path-secret-versions-access-a-secrets-version-using-the-secrets-id)

- `project_id` - (Optional) The ID of the Scaleway Project associated with the secret version.

## Data information

Note: This data source provides you with access to the secret payload, which is encoded in base64.

Keep in mind that this is a sensitive attribute. For more information,
see [Sensitive Data in State](https://developer.hashicorp.com/terraform/language/state/sensitive-data).

~> **Important:**  This property is sensitive and will not be displayed in the Terraform plan, for security reasons.

## Attributes reference

This section lists the attributes that are exported by the scaleway_secret_version data source:

- `description` - (Optional) The description of the secret version (e.g. `my-new-description`).
- `data` - The data payload of the secret version. This is a sensitive attribute containing the secret value. Learn more in the [data section](#data)
- `status` - The status of the secret version.
- `created_at` - The date and time of the secret version's creation in RFC 3339 format.
- `updated_at` - The date and time of the secret version's last update in RFC 3339 format.

Exported attributes are the ones from the `scaleway_secret_version` [resource](../resources/secret_version.md).
