---
page_title: "Instance Server IP Fields Deprecation Guide"
---

# How to modify your Terraform configuration to use the Instance Server IP lists fields

## Instance server `public_ip` and `private_ip` fields replacement

The `public_ip` and `private_ip` computed fields of the `scaleway_instance_server` resource were deprecated since `v2.44.0`
of the provider (`v2.49.0` for `private_ip`) in favor of the `public_ips` and `private_ips` list attributes.
They will be removed in the `v2.61.0` release.

This may be a breaking change for you if you are using those fields in your Terraform configuration, either for output
purposes or in the definition of another resource, for example a load-balancer or an ACL resource.

In this guide, we will take the example of a load-balancer backend resource.

### Single public IP

Here is how we used to declare a load-balancer backend using the `public_ip` attribute.

```hcl
# Declare a public IPv4 addresses
resource "scaleway_instance_ip" "ip_v4" {
  type = "routed_ipv4"
}
# Declare a server with the IP attached
resource "scaleway_instance_server" "server" {
  image = "ubuntu_noble"
  type = "PRO2-S"
  ip_id = scaleway_instance_ip.ip_v4.id
}
# Declare a load-balancer
resource "scaleway_lb" "lb" {
  type = "LB-S"
}
# Declare a backend using the public_ip attribute
resource "scaleway_lb_backend" "backend" {
  lb_id = scaleway_lb.lb.id
  forward_port     = 80
  forward_protocol = "http

  server_ips = [ scaleway_instance_server.server.public_ip ]
}
```

Here is how to declare the same backend using the `public_ips` list attribute.

```hcl
# Declare a backend using the public_ips list attribute
resource "scaleway_lb_backend" "backend" {
  lb_id = scaleway_lb.lb.id
  forward_port     = 80
  forward_protocol = "http"

  server_ips = [ scaleway_instance_server.server.public_ips.0.address ]
}
```

### Private IP

Here is how we used to declare a load-balancer backend using the `private_ip` attribute.

```hcl
# Declare a VPC and a private network
resource "scaleway_vpc" "vpc" {}
resource "scaleway_vpc_private_network" "pn" {
  vpc_id = scaleway_vpc.vpc.id
}
# Declare a server attached to the private network, this will create a private IPv4 and IPv6
resource "scaleway_instance_server" "server" {
  image = "ubuntu_noble"
  type = "PRO2-S"
  private_network {
    pn_id = scaleway_vpc_private_network.pn.id
  }
}
# Declare a load-balancer
resource "scaleway_lb" "lb" {
  type = "LB-S"
}
# Declare a backend with the private IP
resource "scaleway_lb_backend" "backend" {
  lb_id        = scaleway_lb.lb.id
  forward_port = 80
  forward_protocol = "http"

  server_ips = [scaleway_instance_server.server.private_ip]
}
```

Here is how to declare the same backend using the `private_ips` list attribute.

```hcl
# Declare a backend with only the private IPv4 (index 1 in the list)
resource "scaleway_lb_backend" "backend" {
  lb_id        = scaleway_lb.lb.id
  forward_port = 80
  forward_protocol = "http"

  server_ips = [scaleway_instance_server.server.private_ips.1.address]
}
```

```hcl
# Declare a backend with both the private IPv4 and private IPv6
resource "scaleway_lb_backend" "backend" {
  lb_id        = scaleway_lb.lb.id
  forward_port = 80
  forward_protocol = "http"

  server_ips = [
    scaleway_instance_server.server.private_ips.0.address,
    scaleway_instance_server.server.private_ips.1.address,
  ]
}
```

### Multiple public IPs, but only select the IPv4 addresses

This was not achievable with the singular `public_ip` field, but here is how to do it with the `public_ips` field.

**NB**: For private IPs, the logic is the same, just replace `public_ips` by `private_ips`. (Note that the declarations
of private IPs are not shown here)

**NB**: For IPv6 addresses, the logic is the same, just replace `inet` by `inet6`.

```hcl
# Declare 3 public IPv4 addresses
resource "scaleway_instance_ip" "ip_v4" {
  type = "routed_ipv4"
  count = 3
}
# Declare 3 public IPv6 addresses
resource "scaleway_instance_ip" "ip_v6" {
  type = "routed_ipv6"
  count = 3
}
# Declare a server with all IPs attached
resource "scaleway_instance_server" "server" {
  image = "ubuntu_noble"
  type = "PRO2-S"
  ip_ids = concat(
    [ for ip in scaleway_instance_ip.ip_v4 : ip.id ],
    [ for ip in scaleway_instance_ip.ip_v6 : ip.id ],
  )
}
# Declare a load-balancer
resource "scaleway_lb" "lb" {
  type = "LB-S"
}
# Declare a backend with only the IPv4 addresses
resource "scaleway_lb_backend" "backend" {
  lb_id = scaleway_lb.lb.id
  forward_port     = 80
  forward_protocol = "http"

  server_ips = [
    for ip in scaleway_instance_server.server.public_ips : ip.address
    if ip.family == "inet"
  ]
}
```

## Instance server IPv6 fields removal

The fields concerning IPv6 were removed from the `scaleway_instance_server` resource as the information they provided was
no longer accurate.

These fields were:

- `enable_ipv6`, which was false in the state, but is now true for all servers.
- `ipv6_address`, which was null in the state, but can now be retrieved in the `[public|private]_ips.X.address` whether
the IP is public or private.
- `ipv6_gateway`, which was null in the state, but can now be retrieved in the `[public|private]_ips.X.gateway` whether
the IP is public or private.
- `ipv6_prefix_length`, which was 0 in the state.
