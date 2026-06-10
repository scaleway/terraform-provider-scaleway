---
page_title: "Scaleway: scaleway_vpc_route"
subcategory: "VPC"
description: |-
  Lists Scaleway VPC Routes across regions.
---

# Resource: scaleway_vpc_route

For more information, see [the main documentation](https://www.scaleway.com/en/docs/vpc/concepts/).

## Example Usage

```terraform
# List VPC routes across all regions
list "scaleway_vpc_route" "all" {
  provider = scaleway

  config {
    regions = ["*"]
  }
}
```
```terraform
# List VPC routes filtered by nexthop resource type
list "scaleway_vpc_route" "by_nexthop_type" {
  provider = scaleway

  config {
    regions               = ["fr-par"]
    nexthop_resource_type = "instance_private_nic"
  }
}
```
```terraform
# List VPC routes filtered by tag
list "scaleway_vpc_route" "by_tag" {
  provider = scaleway

  config {
    regions = ["*"]
    tags    = ["production"]
  }
}
```
```terraform
# List VPC routes for a specific VPC
list "scaleway_vpc_route" "by_vpc" {
  provider = scaleway

  config {
    regions = ["fr-par"]
    vpc_id  = "11111111-1111-1111-1111-111111111111"
  }
}
```

## Argument Reference

The following arguments can be specified in the `config` block:

- `regions` - (Optional) Regions to filter for. Use `["*"]` to list from all regions.
- `tags` - (Optional) Tags to filter for.
- `vpc_id` - (Optional) Filter for routes belonging to this VPC (regional ID or plain UUID).
- `nexthop_resource_id` - (Optional) Filter for routes with this nexthop resource ID (regional ID or plain UUID).
- `nexthop_private_network_id` - (Optional) Filter for routes with this nexthop private network ID (regional ID or plain UUID).
- `nexthop_vpc_connector_id` - (Optional) Filter for routes with this nexthop VPC connector ID (regional ID or plain UUID).
- `nexthop_resource_type` - (Optional) Filter for routes with this nexthop resource type.
- `is_ipv6` - (Optional) Filter for routes with an IPv6 destination.
- `contains` - (Optional) Filter for routes whose destination is contained in this subnet (CIDR notation).

## Attributes Reference

In addition to the arguments above, each listed VPC route exports the same attributes as the `scaleway_vpc_route` managed resource:

- `id` - The ID of the VPC route.
- `vpc_id` - The ID of the VPC the route belongs to.
- `description` - The description of the route.
- `tags` - The tags associated with the route.
- `destination` - The destination IP or IP range of the route.
- `nexthop_resource_id` - The ID of the nexthop resource.
- `nexthop_private_network_id` - The ID of the nexthop private network.
- `nexthop_vpc_connector_id` - The ID of the nexthop VPC connector.
- `region` - The region of the VPC route.
- `created_at` - The date and time of the creation of the route.
- `updated_at` - The date and time of the last update of the route.
