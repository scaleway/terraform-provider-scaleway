---
page_title: "Scaleway: scaleway_lb"
description: |-
  Gets information about a Load Balancer.
---

# scaleway_lb

Gets information about a Load Balancer.

## Example Usage

```hcl
# Get info by name
data "scaleway_lb" "by_name" {
  name = "foobar"
}

# Get info by ID
data "scaleway_lb" "by_id" {
  lb_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `ip_address` - (Optional) The IP address.
  Only one of `ip_address` and `lb_id` should be specified.

- `lb_id` - (Optional) The ID.
  Only one of `ip_address` and `lb_id` should be specified.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#zones) in which the LB IP exists.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the LB IP is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `reverse` - The reverse domain associated with this IP.

- `lb_id` - The associated load-balance ID if any

- `organization_id` - (Defaults to [provider](../index.md#organization_id) `organization_id`) The ID of the organization the LB IP is associated with.


## Arguments Reference

The following arguments are supported:

- `ip_id` - (Required) The ID of the associated IP. See below.

~> **Important:** Updates to `ip_id` will recreate the load-balancer.

  For now only `LB-S` is available

~> **Important:** Updates to `type` will recreate the load-balancer.


- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the load-balancer should be created.


- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the load-balancer is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the load-balancer.
- `ip_address` -  The load-balance public IP Address
- `organization_id` - The organization ID the load-balancer is associated with.
- `tags` - (Optional) The tags associated with the load-balancers.
- `type` - (Required) The type of the load-balancer.
