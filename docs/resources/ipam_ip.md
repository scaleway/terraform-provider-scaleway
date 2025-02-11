---
subcategory: "IPAM"
page_title: "Scaleway: scaleway_ipam_ip"
---

# Resource: scaleway_ipam_ip

Books and manages IPAM IPs.

For more information about IPAM, see the main [documentation](https://www.scaleway.com/en/docs/vpc/concepts/#ipam).

## Example Usage

### Basic

```terraform
resource "scaleway_vpc" "vpc01" {
  name = "my vpc"
}

resource "scaleway_vpc_private_network" "pn01" {
  vpc_id = scaleway_vpc.vpc01.id
  ipv4_subnet {
    subnet = "172.16.32.0/22"
  }
}

resource "scaleway_ipam_ip" "ip01" {
  source {
    private_network_id = scaleway_vpc_private_network.pn01.id
  }
}
```

### Request a specific IPv4 address

```terraform
resource "scaleway_vpc" "vpc01" {
  name = "my vpc"
}

resource "scaleway_vpc_private_network" "pn01" {
  vpc_id = scaleway_vpc.vpc01.id
  ipv4_subnet {
    subnet = "172.16.32.0/22"
  }
}

resource "scaleway_ipam_ip" "ip01" {
  address = "172.16.32.7"
  source {
    private_network_id = scaleway_vpc_private_network.pn01.id
  }
}
```

### Request an IPv6 address

```terraform
resource "scaleway_vpc" "vpc01" {
  name = "my vpc"
}

resource "scaleway_vpc_private_network" "pn01" {
  vpc_id = scaleway_vpc.vpc01.id
  ipv6_subnets {
    subnet = "fd46:78ab:30b8:177c::/64"
  }
}

resource "scaleway_ipam_ip" "ip01" {
  is_ipv6 = true
  source {
    private_network_id = scaleway_vpc_private_network.pn01.id
  }
}
```

### Book an IP for a custom resource

```terraform
resource "scaleway_vpc" "vpc01" {
  name = "my vpc"
}

resource "scaleway_vpc_private_network" "pn01" {
  vpc_id = scaleway_vpc.vpc01.id
  ipv4_subnet {
    subnet = "172.16.32.0/22"
  }
}

resource "scaleway_ipam_ip" "ip01" {
  address = "172.16.32.7"
  source {
    private_network_id = scaleway_vpc_private_network.pn01.id
  }
  custom_resource {
    mac_address = "bc:24:11:74:d0:6a"
  }
}
```

## Argument Reference

The following arguments are supported:

- `address` - (Optional) Request a specific IP in the specified source pool.

  ~> **Important:** when requesting specific IP addresses, it is best ensure these are created before any other resource in the Private Network. This can be achieved by using `depends_on` relations, or moving the declarations to another Terraform module. Otherwise, other resources may take the requested address first, blocking the whole Terraform setup. Static IPs should be avoided unless necessary, as we cannot guarantee full automation. We recommend to use DNS, or to not request a specific IP.

- `tags` - (Optional) The tags associated with the IP.
- `source` - (Required) The source in which to book the IP.
    - `zonal` - The zone of the IP (if the IP is public and zoned, rather than private and/or regional)
    - `private_network_id` - The Private Network of the IP (if the IP is a private IP).
    - `subnet_id` - The Private Network subnet of the IP (if the IP is a private IP).
- `is_ipv6` - (Optional) Defines whether to request an IPv6 address instead of IPv4.
- `custome_resource` - (Optional) The custom resource to attach to the IP being reserved. An example of a custom resource is a virtual machine hosted on an Elastic Metal server.
    - `mac_address` - The MAC address of the custom resource.
    - `name` - When the resource is in a Private Network, a DNS record is available to resolve the resource name.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) of the IP.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the Project the IP is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the IP in IPAM.
- `resource` - The IP resource.
    - `id` - The ID of the resource that the IP is attached to.
    - `type` - The type of resource the IP is attached to.
    - `name` - The name of the resource the IP is attached to.
    - `mac_address` - The MAC address of the resource the IP is attached to.
- `reverses` - The reverse DNS for this IP.
    - `hostname` The reverse domain name.
    - `address` The IP address corresponding to the hostname.
- `created_at` - Date and time of IP's creation (RFC 3339 format).
- `updated_at` - Date and time of IP's last update (RFC 3339 format).
- `zone` - The zone of the IP.

~> **Important:** IPAM IP IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

## Import

IPAM IPs can be imported using `{region}/{id}`, e.g.

```bash
terraform import scaleway_ipam_ip.ip_demo fr-par/11111111-1111-1111-1111-111111111111
```
