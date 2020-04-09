---
layout: "scaleway"
page_title: "Scaleway: scaleway_lb_beta"
description: |-
  Manages Scaleway Load-Balancers.
---

# scaleway_lb_beta

-> **Note:** This terraform resource is flagged beta and might include breaking change in future releases.

Creates and manages Scaleway Load-Balancers. For more information, see [the documentation](https://developers.scaleway.com/en/products/lb/api).

## Examples
    
### Basic

```hcl
resource "scaleway_lb_ip_beta" "ip" {
}

resource "scaleway_lb_beta" "base" {
  ip_id = scaleway_lb_ip_beta.ip.id
  region      = "fr-par"
  type        = "LB-S"
}
```

## Arguments Reference

The following arguments are supported:

- `ip_id` - (Required) The ID of the associated IP.

~> **Important:** Updates to `ip_id` will recreate the load-balancer.

- `type` - (Required) The type of the load-balancer.  For now only `LB-S` is available

~> **Important:** Updates to `type` will recreate the load-balancer.

- `name` - (Optional) The name of the load-balancer.

- `tags` - (Optional) The tags associated with the load-balancers.

- `region` - (Defaults to [provider](../index.html#region) `region`) The [region](../guides/regions_and_zones.html#regions) in which the load-balancer should be created.

- `organization_id` - (Defaults to [provider](../index.html#organization_id) `organization_id`) The ID of the organization the load-balancer is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the load-balancer.
- `ip_address` -  The load-balance public IP Address


## Import

Load-Balancer can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_lb_beta.lb01 fr-par/11111111-1111-1111-1111-111111111111
```
