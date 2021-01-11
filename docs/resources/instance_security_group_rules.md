---
page_title: "Scaleway: scaleway_instance_security_group_rules"
description: |-
  Manages Scaleway Compute Instance security group rules.
---

# scaleway_instance_security_group_rules

Creates and manages Scaleway Compute Instance security group rules. For more information, see [the documentation](https://developers.scaleway.com/en/products/instance/api/#security-groups-8d7f89).

This resource can be used to externalize rules from a `scaleway_instance_security_group` to solve circular dependency problems. When using this resource do not forget to set `external_rules = true` on the security group.

~> **Warning:** In order to guaranty rules order in a given security group only one scaleway_instance_security_group_rules is allowed per security group.

## Examples

### Basic

```hcl
resource "scaleway_instance_security_group" "sg01" {
  external_rules = true
}

resource "scaleway_instance_security_group_rules" "sgrs01" {
  security_group_id = scaleway_instance_security_group.sg01.id
  inbound_rule {
    action   = "accept"
    port     = 80
    ip_range = "0.0.0.0/0"
  }
}
```

### Simplify your rules using dynamic block and `for_each` loop

You can use [`for_each` syntax](https://www.terraform.io/docs/configuration/meta-arguments/for_each.html) to simplify the definition of your rules.
Let's suppose that your inbound default policy is to drop, but you want to build a list of exceptions to accept.
Create a local containing your exceptions (`locals.trusted`) and use the `for_each` syntax in a [dynamic block](https://www.terraform.io/docs/configuration/expressions/dynamic-blocks.html):

```hcl
resource "scaleway_instance_security_group" "main" {
  description = "test"
  name        = "terraform test"
  inbound_default_policy  = "drop"
  outbound_default_policy = "accept"
}

locals {
  trusted = [
    "1.2.3.4",
    "4.5.6.7",
    "7.8.9.10"
  ]
}

resource "scaleway_instance_security_group_rules" "main" {
  security_group_id       = scaleway_instance_security_group.main.id

  dynamic "inbound_rule" {
    for_each = local.trusted
    content {
      action = "accept"
      ip     = inbound_rule.value
      port   = 80
    }
  }
}
```

You can also use object to assign IP and port in the same time.
In your locals, you can use [objects](https://www.terraform.io/docs/configuration/types.html#structural-types) to encapsulate several values that will be used later on in the loop:

```hcl
resource "scaleway_instance_security_group" "main" {
  description             = "test"
  name                    = "terraform test"
  inbound_default_policy  = "drop"
  outbound_default_policy = "accept"
}

locals {
  trusted = [
    { ip = "1.2.3.4", port = "80" },
    { ip = "5.6.7.8", port = "81" },
    { ip = "9.10.11.12", port = "81" },
  ]
}

resource "scaleway_instance_security_group_rules" "main" {
  security_group_id = scaleway_instance_security_group.main.id

  dynamic "inbound_rule" {
    for_each = local.trusted
    content {
      action = "accept"
      ip     = inbound_rule.value.ip
      port   = inbound_rule.value.port
    }
  }
}
```

## Arguments Reference

The following arguments are supported:

- `security_group_id` - (Required) The ID of the security group.

- `inbound_rule` - (Optional) A list of inbound rule to add to the security group. (Structure is documented below.)

- `outbound_rule` - (Optional) A list of outbound rule to add to the security group. (Structure is documented below.)


The `inbound_rule` and `outbound_rule` block supports:

- `action` - (Required) The action to take when rule match. Possible values are: `accept` or `drop`.

- `protocol`- (Defaults to `TCP`) The protocol this rule apply to. Possible values are: `TCP`, `UDP`, `ICMP` or `ANY`.

- `port`- (Optional) The port this rule apply to. If no port is specified, rule will apply to all port.

- `ip`- (Optional) The ip this rule apply to. If no `ip` nor `ip_range` are specified, rule will apply to all ip. Only one of `ip` and `ip_range` should be specified.

- `ip_range`- (Optional) The ip range (e.g `192.168.1.0/24`) this rule applies to. If no `ip` nor `ip_range` are specified, rule will apply to all ip. Only one of `ip` and `ip_range` should be specified.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the security group.

## Import

Instance security group rules can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_instance_security_group_rules.web fr-par-1/11111111-1111-1111-1111-111111111111
```
