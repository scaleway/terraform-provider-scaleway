---
subcategory: "Secrets"
page_title: "Scaleway: scaleway_secret_version"
---

# Resource: scaleway_secret_version

The `scaleway_secret_version` resource allows you to create and manage secret versions in Scaleway Secret Manager.

Refer to the Secret Manager [product documentation](https://www.scaleway.com/en/docs/identity-and-access-management/secret-manager/) and [API documentation](https://www.scaleway.com/en/developers/api/secret-manager/) for more information.

## Example Usage

### Create a secret and a version

The following commands allow you to:

- create a secret named `foo`
- create a version of this secret containing the `my_new_secret` data

```terraform
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

## Argument Reference

The following arguments are supported:

- `secret_id` - (Required) The ID of the secret associated with the version.
- `data` - (Required) The data payload of the secret version. Must not exceed 64KiB in size (e.g. `my-secret-version-payload`). Find out more on the [data section](/#data-information).
- `description` - (Optional) Description of the secret version (e.g. `my-new-description`).
- `region` - (Defaults to the region specified in the [provider configuration](../index.md#region)). The [region](../guides/regions_and_zones.md#regions) where the resource exists.

### Data

Note: The `data` should be a base64-encoded string when sent from the API. **The provider handles this encoding so you do not need to encode the data yourself.**

Updating `data` will force the creation of a new secret version.

Keep in mind that this is a sensitive attribute. For more information, see [Sensitive Data in State](https://developer.hashicorp.com/terraform/language/state/sensitive-data).

~> **Important:**  This property will not be displayed in the Terraform plan, for security reasons.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `revision` - The revision number of the secret version.
- `status` - The status of the secret version.
- `created_at` - The date and time of the secret version's creation (in RFC 3339 format).
- `updated_at` - The date and time of the secret version's last update (in RFC 3339 format).

## Import

This section explains how to import a secret version using the `{region}/{id}/{revision}` format.

~> **Important:** Keep in mind that if you import with the `latest` revision, you will overwrite the previous version you might have been using.

```bash
terraform import scaleway_secret_version.main fr-par/11111111-1111-1111-1111-111111111111/2
```
