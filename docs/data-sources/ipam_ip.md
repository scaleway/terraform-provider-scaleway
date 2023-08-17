---
subcategory: "IPAM"
page_title: "Scaleway: scaleway_ipam_ip"
---

# scaleway_ipam_ip

Gets information about IP managed by IPAM service. IPAM service is used for dhcp bundled in VPCs' private networks.

## Examples

### Instance Private Network IP

```hcl
# Get Instance IP in a private network
resource "scaleway_instance_private_nic" "nic" {
  server_id = scaleway_instance_server.server.id
  private_network_id = scaleway_vpc_private_network.pn.id
}

# Find server IPv4 using private-nic mac address
data "scaleway_ipam_ip" "ip" {
  mac_address = scaleway_instance_private_nic.nic.mac_address
  type = "ipv4"
}
```

## Argument Reference

- `type` - (Required) The type of IP to search for (ipv4, ipv6).

- `private_network_id` - (Optional) The ID of the private network the IP belong to.

- `resource_id` - (Optional) The ID of the resource that the IP is bound to.

- `mac_address` - (Optional) The Mac Address linked to the IP.

- `region` - (Defaults to [provider](../index.md#zone) `region`) The [region](../guides/regions_and_zones.md#regions) in which the IP exists.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the IP in IPAM
- `address` - The IP address
