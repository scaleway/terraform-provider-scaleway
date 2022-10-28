---
page_title: "Scaleway: scaleway_tem_domain"
description: |-
  Manages Scaleway Transactional Email Domains.
---

# scaleway_tem_domain

Creates and manages Scaleway Transactional Email Domains.
For more information see [the documentation](https://developers.scaleway.com/en/products/registry/api/).

## Examples

### Basic

```hcl
resource "scaleway_tem_domain" "main" {
  name = "example.com"
}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Required) The domain name, must not be used in another Transactional Email Domain.
~> **Important** Updates to `name` will recreate the domain.

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which the domain should be created.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the domain is associated with.

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

- `statistics` - The domain's statistics.

    - `total_count` - The total number of emails matching the request criteria.

    - `new_count` - The number of emails still in the `new` transient state (received from the API, not yet processed).

    - `sending_count` - The number of emails still in the `sending` transient state (received from the API, not yet in their final status).

    - `sent_count` - The number of emails in the final `sent` state (have been delivered to the target mail system).

    - `failed_count` - The number of emails in the final `failed` state (refused by the target mail system with a final error status).

    - `canceled_count` - The number of emails in the final `canceled` state (canceled by customer's request).

## Import

Domains can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_tem_domain.main fr-par/11111111-1111-1111-1111-111111111111
```
