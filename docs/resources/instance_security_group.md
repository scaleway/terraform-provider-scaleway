---
page_title: "Scaleway: scaleway_instance_security_group"
description: |-
  Manages Scaleway Compute Instance security groups.
---

# scaleway_instance_security_group

Creates and manages Scaleway Compute Instance security groups. For more information, see [the documentation](https://developers.scaleway.com/en/products/instance/api/#security-groups-8d7f89).

## Examples

### Basic

```hcl
resource "scaleway_instance_security_group" "allow_all" {
}

resource "scaleway_instance_security_group" "web" {
  inbound_default_policy = "drop" # By default we drop incoming traffic that do not match any inbound_rule
}
```

### Web server with banned IP and restricted internet access

```hcl
resource "scaleway_instance_security_group" "web" {
  inbound_default_policy  = "drop" # By default we drop incoming traffic that do not match any inbound_rule.
  outbound_default_policy = "drop" # By default we drop outgoing traffic that do not match any outbound_rule.
}
```

### Trusted IP for SSH access (using for_each)

If you use terraform >= 0.12.6, you can leverage the [`for_each`](https://www.terraform.io/docs/configuration/resources.html#for_each-multiple-resource-instances-defined-by-a-map-or-set-of-strings) feature with this resource.

```hcl
locals {
  trusted = ["192.168.0.1", "192.168.0.2", "192.168.0.3"]
}

resource "scaleway_instance_security_group" "dummy" {
  inbound_default_policy  = "drop"
  outbound_default_policy = "accept"
}

resource "scaleway_instance_security_group_rules" "main" {
  security_group_id = scaleway_instance_security_group.main.id

  dynamic "inbound_rule" {
    for_each = local.trusted
    content {
      action = "accept"
      ip     = inbound_rule.value.ip
      port   = 22
    }
  }
}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Optional) The name of the security group.

- `description` - (Optional) The description of the security group.

- `stateful` - (Defaults to `true`) A boolean to specify whether the security group should be stateful or not.

- `inbound_default_policy` - (Defaults to `accept`) The default policy on incoming traffic. Possible values are: `accept` or `drop`.

- `outbound_default_policy` - (Defaults to `accept`) The default policy on outgoing traffic. Possible values are: `accept` or `drop`.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the security group should be created.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the security group is associated with.

- `enable_default_security` - Whether to block SMTP on IPv4/IPv6 (Port 25, 465, 587). Set to false will unblock SMTP if your account is authorized to. If your organization is not yet authorized to send SMTP traffic, [open a support ticket](https://console.scaleway.com/support/tickets).

- `tags`- (Optional) The tags of the security group.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the security group.
- `organization_id` - The organization ID the security group is associated with.

## Import

Instance security group can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_instance_security_group.web fr-par-1/11111111-1111-1111-1111-111111111111
```
