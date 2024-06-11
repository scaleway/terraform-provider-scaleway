---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_public_gateway_ip"
---

# scaleway_vpc_public_gateway_ip

Gets information about a Public Gateway public flexible IP address.

For further information, please see the API [documentation](https://www.scaleway.com/en/developers/api/public-gateway/#path-ips-list-ips)

## Example Usage

```hcl
resource "scaleway_vpc_public_gateway_ip" "main" {
}

data "scaleway_vpc_public_gateway_ip" "ip_by_id" {
    ip_id = "${scaleway_vpc_public_gateway_ip.main.id}"
}
```

## Argument Reference

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the Public Gateway IP.

~> **Important:** Public gateway IP IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`
