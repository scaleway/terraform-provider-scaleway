---
layout: "scaleway"
page_title: "Scaleway: scaleway_security_group"
description: |-
  Gets information about a Security Group.
---

# scaleway_security_group

Gets information about a Security Group.

## Example Usage

```hcl
// Get info by security group name
data "scaleway_instance_security_group" "my_key" {
  name  = "my-security-group-name"
}

// Get info by security group id
data "scaleway_instance_security_group" "my_key" {
  security_group_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - The security group name. Only one of `name` and `security_group_id` should be specified.

- `security_group_id` - The security group id. Only one of `name` and `security_group_id` should be specified.

- `zone` - (Defaults to [provider](../index.html#zone) `zone`) The [zone](../guides/regions_and_zones.html#zones) in which the security group should be created.

- `organization_id` - (Defaults to [provider](../index.html#organization_id) `organization_id`) The ID of the organization the server is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the security group.