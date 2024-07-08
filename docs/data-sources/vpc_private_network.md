---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_private_network"
---

# scaleway_vpc_private_network

Gets information about a Private Network.

## Example Usage

```hcl
# Get info by name
data "scaleway_vpc_private_network" "my_name" {
  name = "foobar"
}

# Get info by name and VPC ID
data "scaleway_vpc_private_network" "my_name_and_vpc_id" {
  name   = "foobar"
  vpc_id = "11111111-1111-1111-1111-111111111111"
}

# Get info by IP ID
data "scaleway_vpc_private_network" "my_id" {
  private_network_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) Name of the Private Network. Cannot be used with `private_network_id`.
- `vpc_id` - (Optional) ID of the VPC the Private Network is in. Cannot be used with `private_network_id`.
- `private_network_id` - (Optional) ID of the Private Network. Cannot be used with `name` or `vpc_id`.
- `project_id` - (Optional) The ID of the Project the Private Network is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the Private Network.
- `ipv4_subnet` - The IPv4 subnet associated with the Private Network.
- `ipv6_subnets` - The IPv6 subnets associated with the Private Network.

~> **Important:** Private Networks are [regional](../guides/regions_and_zones.md#resource-ids), which means their IDs are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`
