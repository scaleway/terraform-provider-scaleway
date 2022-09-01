---
page_title: "Scaleway: scaleway_instance_security_group"
description: |-
  Gets information about a Security Group.
---

# scaleway_instance_security_group

Gets information about a Security Group.

## Example Usage

```hcl
# Get info by security group name
data "scaleway_instance_security_group" "my_key" {
  name  = "my-security-group-name"
}

# Get info by security group id
data "scaleway_instance_security_group" "my_key" {
  security_group_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The security group name. Only one of `name` and `security_group_id` should be specified.

- `security_group_id` - (Optional) The security group id. Only one of `name` and `security_group_id` should be specified.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the security group exists.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the security group.

- `organization_id` - The ID of the organization the security group is associated with.

- `project_id` - The ID of the project the security group is associated with.

- `inbound_default_policy` - The default policy on incoming traffic. Possible values are: `accept` or `drop`.

- `outbound_default_policy` - The default policy on outgoing traffic. Possible values are: `accept` or `drop`.