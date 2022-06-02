---
page_title: "Scaleway: scaleway_lb"
description: |-
  Manages Scaleway Load-Balancers.
---

# scaleway_lb

Creates and manages Scaleway Load-Balancers.
For more information, see [the documentation](https://developers.scaleway.com/en/products/lb/zoned_api).

## Examples

### Basic

```hcl
resource "scaleway_lb_ip" "ip" {
  zone = "fr-par-1"
}

resource "scaleway_lb" "base" {
  ip_id  = scaleway_lb_ip.ip.id
  zone   = scaleway_lb_ip.ip.zone
  type   = "LB-S"
}
```

### Multiple configurations

```hcl
### IP for Public Gateway
resource "scaleway_vpc_public_gateway_ip" "main" {
}

### Scaleway Private Network
resource scaleway_vpc_private_network main {
}

### VPC Public Gateway Network
resource "scaleway_vpc_public_gateway" "main" {
    name  = "tf-test-public-gw"
    type  = "VPC-GW-S"
    ip_id = scaleway_vpc_public_gateway_ip.main.id
}

### VPC Public Gateway Network DHCP config
resource "scaleway_vpc_public_gateway_dhcp" "main" {
    subnet = "10.0.0.0/24"
}

### VPC Gateway Network
resource "scaleway_vpc_gateway_network" "main" {
    gateway_id         = scaleway_vpc_public_gateway.main.id
    private_network_id = scaleway_vpc_private_network.main.id
    dhcp_id            = scaleway_vpc_public_gateway_dhcp.main.id
    cleanup_dhcp       = true
    enable_masquerade  = true
}

### Scaleway Instance
resource "scaleway_instance_server" "main" {
    name        = "Scaleway Terraform Provider"
    type        = "DEV1-S"
    image       = "debian_bullseye"
    enable_ipv6 = false

    private_network {
        pn_id = scaleway_vpc_private_network.main.id
    }
}

### IP for LB IP
resource scaleway_lb_ip ip01 {
}

### Scaleway Private Network
resource scaleway_vpc_private_network "static" {
    name = "private network with static config"
}

### Scaleway Load Balancer
resource scaleway_lb lb01 {
    ip_id = scaleway_lb_ip.ip01.id
    name = "test-lb-with-private-network-configs"
    type = "LB-S"

    private_network {
        private_network_id = scaleway_vpc_private_network.static.id
        static_config = ["172.16.0.100", "172.16.0.101"]
    }

    private_network {
        private_network_id = scaleway_vpc_private_network.main.id
        dhcp_config = true
    }

    depends_on = [scaleway_vpc_public_gateway.main]
}
```

## Arguments Reference

The following arguments are supported:

- `ip_id` - (Required) The ID of the associated IP. See below.

~> **Important:** Updates to `ip_id` will recreate the load-balancer.

- `type` - (Required) The type of the load-balancer.

~> **Important:** Updates to `type` will recreate the load-balancer.

- `name` - (Optional) The name of the load-balancer.

- `tags` - (Optional) The tags associated with the load-balancers.

- `release_ip` - (Defaults to false) The release_ip allow release the ip address associated with the load-balancers.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the IP should be reserved.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the load-balancer is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the load-balancer.
- `ip_address` -  The load-balance public IP Address
- `organization_id` - The organization ID the load-balancer is associated with.

~> **Important:** `release_ip` will not be supported. This prevents the destruction of the IP from releasing a LBs.
The `resource_lb_ip` will be the only resource that handles those IPs.
## IP ID

Since v1.15.0, `ip_id` is a required field. This means that now a separate `scaleway_lb_ip` is required.
When importing, the IP needs to be imported as well as the LB.
When upgrading to v1.15.0, you will need to create a new `scaleway_lb_ip` resource and import it.

For instance, if you had the following:

```hcl
resource "scaleway_lb" "base" {
  zone = "fr-par-1"
  type   = "LB-S"
}
```

You will need to update it to:

```hcl
resource "scaleway_lb_ip" "ip" {
}

resource "scaleway_lb" "base" {
  ip_id  = scaleway_lb_ip.ip.id
  zone = "fr-par-1"
  type   = "LB-S"
  release_ip = false
}
```

## Private Network with static config

```hcl
resource scaleway_lb_ip ip01 {
}

resource scaleway_vpc_private_network pnLB01 {
    name = "pn-with-lb-static"
}

resource scaleway_lb lb01 {
    ip_id = scaleway_lb_ip.ip01.id
    name = "test-lb-with-pn-static-2"
    type = "LB-S"
    release_ip = false
    private_network {
        private_network_id = scaleway_vpc_private_network.pnLB01.id
        static_config = ["172.16.0.100", "172.16.0.101"]
    }
}
```

~> **Important:** Updates to `private_network` will recreate the attachment.

- `private_network_id` - (Required) The ID of the Private Network to associate.

- `static_config` - (Optional) Define two local ip address of your choice for each load balancer instance. See below.

- `dhcp_config` - (Optional) Set to true if you want to let DHCP assign IP addresses. See below.

~> **Important:**  Only one of static_config and dhcp_config may be set.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the private network was created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `status` -  The Private Network attachment status

And before running `terraform apply` you will need to import the IP with:

```bash
$ terraform import scaleway_lb_ip.ip fr-par/11111111-1111-1111-1111-111111111111
```

The IP ID can either be found in the console, or you can run:

```bash
$ terraform state show scaleway_lb.base
```

and look for `ip_id`.

## Import

Load-Balancer can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_lb.lb01 fr-par-1/11111111-1111-1111-1111-111111111111
```

Be aware that you will also need to import the `scaleway_lb_ip` resource.
