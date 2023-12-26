---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_private_network"
---

# scaleway_vpc_private_network

Gets information about a private network.

## Example Usage

```hcl
# Get info by name
data "scaleway_vpc_private_network" "my_name" {
  name = "foobar"
}

# Get info by IP ID
data "scaleway_vpc_private_network" "my_id" {
  private_network_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

* `name` - (Optional) Name of the private network. One of `name` and `private_network_id` should be specified.
* `private_network_id` - (Optional) ID of the private network. One of `name` and `private_network_id` should be specified.
- `project_id` - (Optional) The ID of the project the private network is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the private network.
- `ipv4_subnet` - (Optional) The IPv4 subnet associated with the private network.
- `ipv6_subnets` - (Optional) The IPv6 subnets associated with the private network.

~> **Important:** Private networks' IDs are [zoned](../guides/regions_and_zones.md#resource-ids) or [regional](../guides/regions_and_zones.md#resource-ids) if using beta, which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111` or `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111
