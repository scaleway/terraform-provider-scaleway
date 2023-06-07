---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_private_network"
---

# scaleway_vpc_private_network

Gets information about a private network.

## Example Usage

N/A, the usage will be meaningful in the next releases of VPC.

## Argument Reference

* `name` - (Optional) Name of the private network. One of `name` and `private_network_id` should be specified.
* `private_network_id` - (Optional) ID of the private network. One of `name` and `private_network_id` should be specified.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the private network.
- `ipv4_subnet` - (Optional) The IPv4 subnet associated with the private network.
- `ipv6_subnet` - (Optional) The IPv6 subnet(s) associated with the private network.

~> **Important:** Private networks' IDs are [zoned](../guides/regions_and_zones.md#resource-ids) or [regional](../guides/regions_and_zones.md#resource-ids) if using beta, which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111` or `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111
