---
page_title: "Scaleway: security_group"
description: |-
  Manages Scaleway security groups.
---

# scaleway_security_group

**DEPRECATED**: This resource is deprecated and will be removed in `v2.0+`.
Please use `scaleway_instance_security_group` instead.

Provides security groups. This allows security groups to be created, updated and deleted.
For additional details please refer to [API documentation](https://developer.scaleway.com/#security-groups).

## Example Usage

```hcl
resource "scaleway_security_group" "test" {
  name                    = "test"
  description             = "test"
  enable_default_security = true
  stateful                = true
  inbound_default_policy  = "accept"
  outbound_default_policy = "drop"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) name of security group
* `description` - (Required) description of security group
* `enable_default_security` - (Optional) default: true. Add default security group rules
* `stateful` - (Optional) default: false. Mark the security group as stateful. Note that stateful security groups can not be associated with bare metal servers
* `inbound_default_policy` - (Optional) default policy for inbound traffic. Can be one of accept or drop
* `outbound_default_policy` - (Optional) default policy for outbound traffic. Can be one of accept or drop

Field `name`, `description` are editable.

## Attributes Reference

The following attributes are exported:

* `id` - id of the new resource

## Import

Instances can be imported using the `id`, e.g.

```
$ terraform import scaleway_security_group.test 5faef9cd-ea9b-4a63-9171-9e26bec03dbc
```
