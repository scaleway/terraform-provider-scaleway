---
page_title: "Scaleway: scaleway_secret_version"
subcategory: "Secret"
description: |-
  Lists Scaleway Secret Versions.
---

# Resource: scaleway_secret_version

Lists Scaleway Secret Versions.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/secret-manager/concepts/).

## Example Usage

```terraform
// List all versions of all secrets in all regions
list "scaleway_secret_version" "all_secrets" {
  provider = scaleway

  config {
    regions    = ["*"]
    secret_ids = ["*"]
  }
}
```

```terraform
// List all versions of a secret
list "scaleway_secret_version" "all" {
  provider = scaleway

  config {
    secret_ids = [scaleway_secret.my_secret.id]
  }
}
```

```terraform
// List enabled versions of a secret
list "scaleway_secret_version" "enabled" {
  provider = scaleway

  config {
    secret_ids = [scaleway_secret.my_secret.id]
    status     = ["enabled"]
  }
}
```


## Argument Reference

The following arguments can be specified in the `config` block:

- `regions` - (Optional) Regions to target. Use '*' to list from all regions. If not specified, the provider default region is used.
- `project_ids` - (Optional) Project IDs to filter for. Use '*' to list across all projects. If not specified, the provider default project ID is used.
- `organization_id` - (Optional) Organization ID to filter for.
- `secret_ids` - (Required) IDs of the secrets to list versions for. Use '*' to list versions from all secrets. If empty, returns an error.
- `status` - (Optional) Filter by status. Possible values: `enabled`, `disabled`, `scheduled_for_deletion`, `deleted`.

## Attributes Reference

In addition to the arguments above, the following attributes are exported for each Secret Version:

- `secret_id` - ID of the secret.
- `revision` - The revision number of the version.
- `description` - Description of the version.
- `status` - Status of the version.
- `created_at` - Date and time of the version's creation (RFC 3339 format).
- `updated_at` - Date and time of the version's last update (RFC 3339 format).
- `region` - Region of the version.
