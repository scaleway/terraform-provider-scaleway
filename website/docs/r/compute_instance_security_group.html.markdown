---
layout: "scaleway"
page_title: "Scaleway: scaleway_compute_instance_server"
sidebar_current: "docs-scaleway-resource-compute-instance-server"
description: |-
  Manages Scaleway Compute Instance servers.
---

# scaleway_compute_instance_server

Creates and manages Scaleway Compute Instance security groups. For more information, see [the documentation](https://developers.scaleway.com/en/products/instance/api/#security-groups-8d7f89).

## Examples

### Basic

```hcl
resource "scaleway_compute_instance_security_group" "allow_all" {
}

resource "scaleway_compute_instance_security_group" "web" {
  inbound_default_policy = "drop" # By default we drop incomming trafic that do not match any inbound_rule
  
  
  inbound_rule {
    action = "accept"
    port = 22
    ip = "212.47.225.64"
  }
  
  inbound_rule {
    action = "accept"
    port = 80
  }
}
```

### Web server with banned IP and restricted internet access

```hcl
resource "scaleway_compute_instance_security_group" "web" {
  inbound_default_policy = "drop" # By default we drop incomming trafic that do not match any inbound_rule.
  inbound_default_policy = "drop" # By default we drop outgoing trafic that do not match any outbound_rule.
  
  inbound_rule {
    action = "drop"
    ip = "1.1.1.1" # Banned IP
  }
  
  inbound_rule {
    action = "accept"
    port = 22
    ip = "212.47.225.64"
  }
  
  inbound_rule {
    action = "accept"
    port = 443
  }
  
  outbound_rule {
    action = "accept"
    ip = "8.8.8.8" # Only allow outgoing conection to this IP.
  }
}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Optional) The name of the security group.

- `description` - (Optional) The description of the security group.

- `inbound_default_policy` - (Defaults to `accept`) The default policy on incoming traffic. Possible values are: `accept` or `drop`.

- `outbound_default_policy` - (Defaults to `accept`) The default policy on outgoing traffic. Possible values are: `accept` or `drop`.

- `inbound_rule` - (Optional) A list of inbound rule to add to the security group. (Structure is documented below.)

- `outbound_rule` - (Optional) A list of outbound rule to add to the security group. (Structure is documented below.)

- `zone` - (Defaults to [provider](../index.html#zone) `zone`) The [zone](../guides/regions_and_zones.html#zones) in which the server should be created.

- `project_id` - (Defaults to [provider](../index.html#project_id) `project_id`) The ID of the project the server is associated with.


The `inbound_rule` and `outbound_rule` block supports:

- `action` - (Required) The action to take when rule match. Possible values are: `accept` or `drop`.

- `protocol`- (Optional) The protocol this rule apply to. Possible values are: `TCP` or `UDP`, `ICMP`.

- `port`- (Optional) The port this rule apply to. If no port is specified, rule will apply to all port.

- `ip`- (Optional) The ip this rule apply to. If no `ip` nor `ip_range` are specified, rule will apply to all ip. Only one of `ip` and `ip_range` should be specified.

- `ip_range`- (Optional) The ip range (e.g `192.168.1.0/24`) this rule apply to. If no `ip` nor `ip_range` are specified, rule will apply to all ip. Only one of `ip` and `ip_range` should be specified.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the server.

## Import

Instance security group can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_compute_instance_security_group.web fr-par-1/11111111-1111-1111-1111-111111111111
```
