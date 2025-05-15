---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_routes"
---

# scaleway_vpc_routes

Gets information about multiple VPC routes.

## Example Usage

```hcl
resource scaleway_vpc vpc01 {
  name           = "tf-vpc-route"
  enable_routing = true
}

resource scaleway_vpc_private_network pn01 {
  name   = "tf-pn-route"
  vpc_id = scaleway_vpc.vpc01.id
}

resource scaleway_vpc_private_network pn02 {
  name   = "tf-pn_route-2"
  vpc_id = scaleway_vpc.vpc01.id
}

# Find routes with a matching VPC ID
data "scaleway_vpc_routes" "routes_by_vpc_id" {
  vpc_id = scaleway_vpc.vpc01.id
}

# Find routes with a matching next hop private network ID
data "scaleway_vpc_routes" "routes_by_pn_id" {
  vpc_id                     = scaleway_vpc.vpc01.id
  nexthop_private_network_id = scaleway_vpc_private_network.pn01.id
}

# Find routes with an IPv6 destination 
data "scaleway_vpc_routes" "routes_by_pn_id" {
  vpc_id  = scaleway_vpc.vpc01.id
  is_ipv6 = true
}

# Find routes with a nexthop resource type
data "scaleway_vpc_routes" "routes_by_pn_id" {
  vpc_id                = scaleway_vpc.vpc01.id
  nexthop_resource_type = "vpc_gateway_network"
}
```

## Argument Reference

- `vpc_id` - (Optional) The VPC ID to filter for. routes with a similar VPC ID are listed.

- `nexthop_resource_id` - (Optional) The next hop resource ID to filter for. routes with a similar next hop resource ID are listed.

- `nexthop_private_network_id` - (Optional) The next hop private network ID to filter for. routes with a similar next hop private network ID are listed.

- `nexthop_resource_type` - (Optional) The next hop resource type to filter for. routes with a similar next hop resource type are listed.

- `is_ipv6` - (Optional) Routes with an IPv6 destination will be listed.

- `tags` - (Optional) List of tags to filter for. routes with these exact tags are listed.

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which the routes exist.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `routes` - List of retrieved routes
    - `id` - The ID of the route.
      ~> **Important:** route IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111
    - `created_at` - The date on which the route was created (RFC 3339 format).
    - `destination` - The destination IP or IP range of the route.
    - `description` - The description of the route.
    - `nexthop_ip` - The IP of the route's next hop.
    - `nexthop_name` - The name of the route's next hop.
