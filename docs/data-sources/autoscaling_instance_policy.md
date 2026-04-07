---
subcategory: "Autoscaling"
page_title: "Scaleway: scaleway_autoscaling_instance_policy"
---

# scaleway_autoscaling_instance_policy

Gets information about an Autoscaling Instance policy.

## Example Usage

```hcl
# Get info by name (instance_group_id is required when using name)
data "scaleway_autoscaling_instance_policy" "by_name" {
  name              = "my-instance-policy"
  instance_group_id = scaleway_autoscaling_instance_group.main.id
}

# Get info by ID
data "scaleway_autoscaling_instance_policy" "by_id" {
  instance_policy_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The name of the Instance policy. Only one of `name` and `instance_policy_id` should be specified. When using `name`, `instance_group_id` is required.
- `instance_group_id` - (Optional) The ID of the Instance group the policy belongs to. Required when looking up by `name`.
- `instance_policy_id` - (Optional) The ID of the Instance policy. Only one of `name` and `instance_policy_id` should be specified.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the Instance policy exists.

## Attributes Reference

Exported attributes are the ones from `scaleway_autoscaling_instance_policy` [resource](../resources/autoscaling_instance_policy.md).
