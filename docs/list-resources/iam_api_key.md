---
page_title: "Scaleway: scaleway_iam_api_key"
subcategory: "IAM"
description: |-
  Lists Scaleway IAM API Keys.
---

# Resource: scaleway_iam_api_key

Lists Scaleway IAM API Keys.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/iam/concepts/).

## Example Usage

```terraform
// List all API keys in an organization
list "scaleway_iam_api_key" "all" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
  }
}
```

```terraform
// List API keys filtered by description
list "scaleway_iam_api_key" "by_description" {
  provider = scaleway

  config {
    description = "production"
  }
}
```

```terraform
// List API keys filtered by editable status
list "scaleway_iam_api_key" "by_editable" {
  provider = scaleway

  config {
    editable = true
  }
}
```


## Argument Reference

The following arguments can be specified in the `config` block:

- `organization_id` - (Optional) Organization ID to filter for. If not specified, the provider default organization is used.
- `editable` - (Optional) Filter by editable status.
- `expired` - (Optional) Filter by expired status.
- `description` - (Optional) Filter by description.
- `bearer_id` - (Optional) Filter by bearer ID.
- `bearer_type` - (Optional) Filter by type of bearer (user or application).
- `access_keys` - (Optional) Filter by a list of access keys.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each API Key:

- `id` - The access key of the API key.
- `description` - The description of the iam api key.
- `created_at` - The date and time of the creation of the iam api key.
- `updated_at` - The date and time of the last update of the iam api key.
- `expires_at` - The date and time of the expiration of the iam api key.
- `access_key` - The access key of the iam api key.
- `application_id` - ID of the application attached to the api key.
- `user_id` - ID of the user attached to the api key.
- `editable` - Whether or not the iam api key is editable.
- `creation_ip` - The IPv4 Address of the device which created the API key.
- `default_project_id` - The default project ID associated with the API key.
