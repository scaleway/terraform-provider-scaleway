---
page_title: "Scaleway: scaleway_lb_private_network"
description: |-
Manages Scaleway Load-Balancers private networks
---

# scaleway_lb_private_network

Creates and manages Scaleway attach/detach private networks to Load-Balancers.
For more information, see [the documentation](https://developers.scaleway.com/en/products/lb/zoned_api/#post-d4b30).

## Examples

### Basic

```hcl
resource "scaleway_vpc_private_network" "pn01" {
  name = "test-lb-pn"
}

resource "scaleway_lb_ip" "ip01" {}

resource "scaleway_lb" "lb01" {
  ip_id = scaleway_lb_ip.ip01.id
  name = "test-lb"
  type = "lb-s"
  release_ip = true
}

resource "scaleway_lb_private_network" "lb01pn01" {
  lb_id = scaleway_lb.lb01.id
  private_network_id = scaleway_vpc_private_network.pn01.id
  static_config = ["172.16.0.100", "172.16.0.101"]
}
```

## Arguments Reference

The following arguments are supported:

- `lb_id` - (Required) The ID of the load-balancer to associate.

- `private_network_id` - (Required) The ID of the Private Network to associate.

- `static_config` - (Required) Define two local ip address of your choice for each load balancer instance. See below.

- `dhcp_config` - (Required) Set to true if you want to let DHCP assign IP addresses. See below.

~> **Important:**  Only one of static_config and dhcp_config may be set.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the IP should be reserved.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the load-balancer is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `status` -  The Private Network attachment status

## Import

Load-Balancer private network can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_lb_private_network.lb01pn01 fr-par/11111111-1111-1111-1111-111111111111
```

The attachments can either be found in the console
