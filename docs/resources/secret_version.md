---
subcategory: "Secrets"
page_title: "Scaleway: scaleway_secret_version"
---

# Resource: scaleway_secret_version

The `scaleway_secret_version` resource allows you to create and manage secret versions in Scaleway Secret Manager.

Refer to the Secret Manager [product documentation](https://www.scaleway.com/en/docs/secret-manager/) and [API documentation](https://www.scaleway.com/en/developers/api/secret-manager/) for more information.

-> **Security Best Practice:**
For enhanced security, we recommend using the [`data_wo` write-only argument](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) instead of the regular `data` argument. This ensures your sensitive data is never stored in Terraform state files, providing superior protection against accidental exposure. Write-Only arguments are supported in Terraform 1.11.0 and later.

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
- `data` - (Optional) The data payload of the secret version, encoded as defined by the `type` attribute. Must not exceed 64KiB in size. Find out more on the [data section](#data).
- `data_wo` - (Optional) The data payload of your secret version in [write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only) mode, encoded as defined by the `type` attribute. Must not exceed 64KiB in size. Find out more on the [data section](#data).
- `type` - (Optional, Defaults to `string`) The encoding of the `data` or `data_wo` attribute. Can be either `string` or `base64`. Find out more on the [data section](#data).
- `data_wo_version` - (Optional) The version of the [write-only](https://developer.hashicorp.com/terraform/language/manage-sensitive-data/write-only) data. To update the `data_wo`, you must also update the `data_wo_version`.
- `description` - (Optional) Description of the secret version (e.g. `my-new-description`).
- `region` - (Optional, Computed, Defaults to the region specified in the [provider configuration](../index.md#arguments-reference)). The [region](../guides/regions_and_zones.md#regions) where the resource exists.

### Data

Only one of `data` or `data_wo` should be specified. If both are provided, Terraform will return an error.

The encoding of the payload is controlled by the `type` attribute:

- `string` (default): `data` (or `data_wo`) is used as-is. This is the recommended option for text secrets.
- `base64`: `data` (or `data_wo`) must be the base64 encoding of the payload. The provider decodes it before creating the secret version. This is required for data that is not a valid UTF-8 string (e.g. binary files).

The maximum size for the payload is 64KiB.

#### data
Updating `data` (or changing its `type`) will force the creation of a new secret version.

The value of `data` stored in the Terraform state is the value provided in the configuration (the raw string, or its base64 encoding when `type = "base64"`). Keep in mind that this is a sensitive attribute. For more information, see [Sensitive Data in State](https://developer.hashicorp.com/terraform/language/state/sensitive-data).

~> **Breaking change for existing secret versions:**
Before this version, the state stored `data` as its **base64 encoding** (or verbatim if it was already a valid base64 string), while the API stored the configured value itself. The state now stores the value exactly as provided in the configuration.

Existing states keep their legacy value (with `type = "string"` added), so the upgrade itself recreates nothing. However, if your configuration no longer matches the legacy value — which is always the case when your original `data` was not a valid base64 string — the first plan after the upgrade recreates the version, which then stores the value as declared.

To avoid the one-time recreation, you can fix the state manually before the first apply: set the resource's `data` to the value from your configuration.

```bash
terraform state pull >state.json
# edit the resource's attributes.data to the value from your configuration (e.g. "aGVsbG8=" to "hello")
terraform state push state.json
```

After such a recreation, the payload in the API and the `scaleway_secret_version.X.data` expression return the raw configured value. If you reference it in your configuration (for example to feed another resource):

- You previously relied on the **base64** value → wrap it: `base64encode(scaleway_secret_version.X.data)`.
- You used `base64decode(scaleway_secret_version.X.data)` to obtain the raw secret → remove the `base64decode(...)` and use `scaleway_secret_version.X.data` directly.

#### type
The encoding of the `data` or `data_wo` attribute. Defaults to `string`. Set it to `base64` to store data that is not a valid UTF-8 string (e.g. binary files): the provided value is then expected to be the base64 encoding of the payload and is decoded before being sent to the API. Changing `type` will force the creation of a new secret version.

#### data_wo
`data_wo` is a [write-only](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/guides/using-write-only-arguments) argument, ideal for sensitive secrets that should not be stored in Terraform state. As such, its value is ephemeral and **will not be stored in the state** or displayed in plans. It follows the same `type` encoding rules as `data`.

To update the `data_wo` and trigger a new version creation, you must also update the `data_wo_version` attribute.

~> **Important:** This property will not be displayed in the Terraform plan, for security reasons.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `revision` - The revision number of the secret version.
- `status` - The status of the secret version.
- `created_at` - The date and time of the secret version's creation (in RFC 3339 format).
- `updated_at` - The date and time of the secret version's last update (in RFC 3339 format).
- `srn` - The Scaleway Resource Name (SRN) of the secret version.

## Import

This section explains how to import a secret version using the `{region}/{id}/{revision}` format.

~> **Important:** Keep in mind that if you import with the `latest` revision, you will overwrite the previous version you might have been using.

```bash
terraform import scaleway_secret_version.main fr-par/11111111-1111-1111-1111-111111111111/2
```
