---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_public_gateway_dhcp"
---

# scaleway_vpc_public_gateway_dhcp  

~> **Important:**  The data source `scaleway_vpc_public_gateway_dhcp` has been deprecated and will no longer be supported.
In 2023, DHCP functionality was moved from Public Gateways to Private Networks, DHCP resources are now no longer needed.
For more information, please refer to the [dedicated guide](../guides/migration_guide_vpcgw_v2.md).

Gets information about a Public Gateway DHCP configuration.

## Example Usage

```hcl
resource "scaleway_vpc_public_gateway_dhcp" "main" {
  subnet = "192.168.0.0/24"
}

data "scaleway_vpc_public_gateway_dhcp" "dhcp_by_id" {
    dhcp_id = "${scaleway_vpc_public_gateway_dhcp.main.id}"
}
```

## Argument Reference


## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the Public Gateway DHCP configuration.

~> **Important:** Public Gateway DHCP configuration IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

