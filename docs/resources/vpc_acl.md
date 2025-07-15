---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_acl"
---

# Resource: scaleway_vpc_acl

Creates and manages Scaleway VPC ACLs.

## Example Usage

### Basic

```terraform
resource "scaleway_vpc" "vpc01" {
  name = "tf-vpc-acl"
}

resource "scaleway_vpc_acl" "acl01" {
  vpc_id  = scaleway_vpc.vpc01.id
  is_ipv6 = false
  rules {
    protocol      = "TCP"
    src_port_low  = 0
    src_port_high = 0
    dst_port_low  = 80
    dst_port_high = 80
    source        = "0.0.0.0/0"
    destination   = "0.0.0.0/0"
    description   = "Allow HTTP traffic from any source"
    action        = "accept"
  }
  default_policy = "drop"
}
```

## Argument Reference

The following arguments are supported:

- `vpc_id` - (Required) The VPC ID the ACL belongs to.
- `default_policy` - (Optional. Defaults to `accept`) The action to take for packets which do not match any rules.
- `is_ipv6` - (Optional) Defines whether this set of ACL rules is for IPv6 (false = IPv4). Each Network ACL can have rules for only one IP type.
- `rules` - (Optional) The list of Network ACL rules.
    - `protocol` - (Optional) The protocol to which this rule applies. Default value: ANY.
    - `source` - (Optional) The Source IP range to which this rule applies (CIDR notation with subnet mask).
    - `src_port_low` - (Optional) The starting port of the source port range to which this rule applies (inclusive).
    - `src_port_high` - (Optional) The ending port of the source port range to which this rule applies (inclusive).
    - `destination` - (Optional) The destination IP range to which this rule applies (CIDR notation with subnet mask).
    - `dst_port_low` - (Optional) The starting port of the destination port range to which this rule applies (inclusive).
    - `dst_port_high` - (Optional) The ending port of the destination port range to which this rule applies (inclusive).
    - `action` - (Optional) The policy to apply to the packet.
    - `description` - (Optional) The rule description.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) of the ACL.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the ACL.

~> **Important:** ACLs' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111

## Import

ACLs can be imported using `{region}/{id}`, e.g.

```bash
terraform import scaleway_vpc_acl.main fr-par/11111111-1111-1111-1111-111111111111
```
