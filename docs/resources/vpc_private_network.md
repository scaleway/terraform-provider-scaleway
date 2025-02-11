---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_private_network"
---

# Resource: scaleway_vpc_private_network

Creates and manages Scaleway VPC Private Networks.
For more information, see [the API documentation](https://www.scaleway.com/en/developers/api/vpc/#private-networks-ac2df4).

## Example Usage

### Basic

```terraform
resource "scaleway_vpc_private_network" "pn_priv" {
    name = "subnet_demo"
    tags = ["demo", "terraform"]
}
```

### With subnets

```terraform
resource "scaleway_vpc_private_network" "pn_priv" {
    name = "subnet_demo"
    tags = ["demo", "terraform"]
    
    ipv4_subnet {
      subnet = "192.168.0.0/24"
    }
    ipv6_subnets {
      subnet = "fd46:78ab:30b8:177c::/64"
    }
    ipv6_subnets {
      subnet = "fd46:78ab:30b8:c7df::/64"
    }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Optional) The name of the Private Network. If not provided, it will be randomly generated.
- `tags` - (Optional) The tags associated with the Private Network.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the Project the private network is associated with.
- `ipv4_subnet` - (Optional) The IPv4 subnet to associate with the Private Network.
    - `subnet` - (Optional) The subnet CIDR.
- `ipv6_subnets` - (Optional) The IPv6 subnets to associate with the private network.
    - `subnet` - (Optional) The subnet CIDR.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) of the Private Network.
- `vpc_id` - (Optional) The VPC in which to create the Private Network.
- `is_regional` - (Deprecated) Private Networks are now all necessarily regional.
- `zone` - (Deprecated) Use `region` instead.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Private Network.
- `created_at` - The date and time of the creation of the Private Network (RFC 3339 format).
- `updated_at` - The date and time of the creation of the Private Network (RFC 3339 format).
- `ipv4_subnet` - The IPv4 subnet associated with the Private Network.
    - `subnet` - The subnet CIDR.
    - `id` - The subnet ID.
    - `created_at` - The date and time of the creation of the subnet.
    - `updated_at` - The date and time of the last update of the subnet.
    - `address` - The network address of the subnet in dotted decimal notation, e.g., '192.168.0.0' for a '192.168.0.0/24' subnet.
    - `subnet_mask` - The subnet mask expressed in dotted decimal notation, e.g., '255.255.255.0' for a /24 subnet
    - `prefix_length` - The length of the network prefix, e.g., 24 for a 255.255.255.0 mask.
- `ipv6_subnets` - The IPv6 subnets associated with the Private Network.
    - `subnet` - The subnet CIDR.
    - `id` - The subnet ID.
    - `created_at` - The date and time of the creation of the subnet.
    - `updated_at` - The date and time of the last update of the subnet.
    - `address` - The network address of the subnet in hexadecimal notation, e.g., '2001:db8::' for a '2001:db8::/64' subnet.
    - `prefix_length` - The length of the network prefix, e.g., 64 for a 'ffff:ffff:ffff:ffff::' mask.

~> **Important:** Private networks' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form of `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111

- `organization_id` - The Organization ID the Private Network is associated with.

## Import

Private Networks can be imported using `{region}/{id}`, e.g.

```bash
terraform import scaleway_vpc_private_network.main fr-par/11111111-1111-1111-1111-111111111111
```
