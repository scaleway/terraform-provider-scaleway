---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_route"
---

# scaleway_vpc_route

Gets information about a VPC route.

## Example Usage

```terraform
# Get info by filters
data "scaleway_vpc_route" "by_filters" {
  vpc_id = scaleway_vpc.my_vpc.id
  tags   = ["my-tag"]
}
```

```terraform
# Get info by route ID
data "scaleway_vpc_route" "by_id" {
  route_id = "fr-par/11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

One of `route_id` or filter arguments must be specified.

- `route_id` - (Optional) The ID of the route.

The following filter arguments are supported (cannot be used with `route_id`):

- `vpc_id` - (Optional) The VPC ID to filter for. Only routes within this VPC will be returned.

- `nexthop_resource_id` - (Optional) The next hop resource ID to filter for. Only routes with a matching next hop resource ID will be returned.

- `nexthop_private_network_id` - (Optional) The next hop private network ID to filter for. Only routes with a matching next hop private network ID will be returned.

- `nexthop_resource_type` - (Optional) The next hop resource type to filter for. Only routes with a matching next hop resource type will be returned.

- `is_ipv6` - (Optional) If true, only routes with an IPv6 destination will be returned.

- `tags` - (Optional) List of tags to filter for. Only routes with these exact tags will be returned.

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which the route exists.

~> **Note:** When using filter arguments, the filters must match exactly one route. If zero or multiple routes are found, an error will be returned.

## Attributes Reference

Exported attributes are the ones from `scaleway_vpc_route` [resource](../resources/vpc_route.md).
