---
page_title: "Scaleway: scaleway_domain_zone"
description: |-
  Manages Scaleway Domain zones.
---

# scaleway_domain_zone

Creates and manages Scaleway Domain zone.  
For more information, see [the documentation](https://www.scaleway.com/en/docs/scaleway-dns/).

## Examples


```hcl
resource "scaleway_domain_zone" "test" {
  domain    = "scaleway-terraform.com"
  subdomain = "test"
}
```

## Arguments Reference

The following arguments are supported:

- `domain` - (Required) The domain where the DNS zone will be created.

- `subdomain` - (Required) The subdomain(zone name) to create in the domain.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the domain is associated with.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `ns` - NameServer list for zone.

- `ns_default` - NameServer default list for zone.

- `ns_master` - NameServer master list for zone.

- `status` - The domain zone status.

- `message` - Message

- `updated_at` - The date and time of the last update of the DNS zone.

## Import

Zone can be imported using the `{subdomain}.{domain}`, e.g.

```bash
$ terraform import scaleway_domain_zone.test test.scaleway-terraform.com
```
