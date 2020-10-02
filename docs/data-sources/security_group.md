---
layout: "scaleway"
page_title: "Scaleway: scaleway_security_group"
description: |-
  Gets information about a Security Group.
---

# scaleway_security_group

**DEPRECATED**: This resource is deprecated and will be removed in `v2.0+`.
Please use `scaleway_instance_security_group` instead.

Gets information about a Security Group.

## Example Usage

```hcl
data "scaleway_security_group" "test" {
  name = "my-security-group"
}
```

## Argument Reference

* `name` - (Required) Exact name of desired Security Group

## Attributes Reference

`id` is set to the ID of the found Image. In addition, the following attributes
are exported:

* `description` - description of the security group
* `enable_default_security` - have default security group rules been added to this security group?
