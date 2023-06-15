---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_private_network"
---

# scaleway_vpc_private_network

Gets information about a private network.

## Example Usage

### Basic

```hcl
# Get info by name
data "scaleway_vpc_private_network" "my_name" {
  name = "foobar"
}

# Get info by IP ID
data "scaleway_vpc_private_network" "my_id" {
  private_network_id_id = "11111111-1111-1111-1111-111111111111"
}
```

### Regional

```hcl
# Get info by name
data "scaleway_vpc_private_network" "my_name" {
  name = "foobar"
  is_regional = true
}

# Get info by IP ID
data "scaleway_vpc_private_network" "my_id" {
  private_network_id_id = "11111111-1111-1111-1111-111111111111"
  is_regional = true
}
```

## Argument Reference

* `name` - (Optional) Name of the private network. One of `name` and `private_network_id` should be specified.
* `private_network_id` - (Optional) ID of the private network. One of `name` and `private_network_id` should be specified.
* `is_regional` - (Optional) Whether this is a regional or zonal private network.

## Attributes Reference

See the [VPC Private Network Resource](../resources/vpc_private_network.md) for details on the returned attributes - they are identical.

~> **Important:** Private networks' IDs are [zoned](../guides/regions_and_zones.md#resource-ids) or [regional](../guides/regions_and_zones.md#resource-ids) if using beta, which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111` or `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111
