---
page_title: "Scaleway: scaleway_tem_domain"
description: |-
  Gets information about a transactional email domain.
---

# scaleway_tem_domain

Gets information about a transactional email domain.

## Example Usage

```hcl
// Get info by domain name
data "scaleway_tem_domain" "my_domain" {
  name = "example.com"
}

// Get info by domain ID
data "scaleway_tem_domain" "my_domain" {
  id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The domain name.
  Only one of `name` and `id` should be specified.

- `id` - (Optional) The domain id.
  Only one of `name` and `id` should be specified.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the domain exists.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the domain is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the Transaction Email Domain.

- `status` - The status of the Transaction Email Domain.

- `created_at` - The date and time of the Transaction Email Domain's creation (RFC 3339 format).

- `next_check_at` - The date and time of the next scheduled check (RFC 3339 format).

- `last_valid_at` - The date and time the domain was last found to be valid (RFC 3339 format).

- `revoked_at` - The date and time of the revocation of the domain (RFC 3339 format).

- `last_error` - The error message if the last check failed.

- `spf_config` - The snippet of the SPF record that should be registered in the DNS zone.

- `dkim_config` - The DKIM public key, as should be recorded in the DNS zone.
