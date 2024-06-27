---
subcategory: "IAM"
page_title: "Scaleway: scaleway_iam_api_key"
---

# Resource: scaleway_iam_api_key

Creates and manages Scaleway API Keys. For more information, refer to the [IAM API documentation](https://www.scaleway.com/en/developers/api/iam/#api-keys-3665ae).

## Example Usage

### With application

```terraform
resource "scaleway_iam_application" "ci_cd" {
  name = "My application"
}

resource "scaleway_iam_api_key" "main" {
  application_id = scaleway_iam_application.main.id
  description    = "a description"
}
```

### With user

```terraform
resource "scaleway_iam_user" "main" {
  email = "test@test.com"
}

resource "scaleway_iam_api_key" "main" {
  user_id = scaleway_iam_user.main.id
  description    = "a description"
}
```

## Argument Reference

The following arguments are supported:

- `description`: (Optional) The description of the API key.
- `application_id`: (Optional) ID of the application attached to the API key.
- `user_id` - (Optional) ID of the user attached to the API key.
  -> **Note** You must specify at least one: `application_id` and/or `user_id`.
- `expires_at` - (Optional) The date and time of the expiration of the IAM API key. Please note that in case of any changes,
  the resource will be recreated.
- `default_project_id` - (Optional) The default Project ID to use with Object Storage.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the API key, which is the access key.
- `created_at` - The date and time of the creation of the IAM API key.
- `updated_at` - The date and time of the last update of the IAM API key.
- `editable` - Whether the IAM API key is editable.
- `access_key` - The access key of the IAM API key.
- `secret_key`: The secret Key of the IAM API key.
- `creation_ip` - The IP Address of the device which created the API key.

## Import

Api keys can be imported using the `{id}`, e.g.

```bash
$ terraform import scaleway_iam_api_key.main 11111111111111111111
```
