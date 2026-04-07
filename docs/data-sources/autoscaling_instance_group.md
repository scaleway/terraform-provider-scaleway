---
subcategory: "Autoscaling"
page_title: "Scaleway: scaleway_autoscaling_instance_group"
---

# scaleway_autoscaling_instance_group

Gets information about an Autoscaling Instance group.

## Example Usage

```hcl
# Get info by name
data "scaleway_autoscaling_instance_group" "by_name" {
  name = "my-instance-group"
}

# Get info by ID
data "scaleway_autoscaling_instance_group" "by_id" {
  instance_group_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The name of the Instance group. Only one of `name` and `instance_group_id` should be specified.
- `instance_group_id` - (Optional) The ID of the Instance group. Only one of `name` and `instance_group_id` should be specified.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the Instance group exists.

## Attributes Reference

Exported attributes are the ones from `scaleway_autoscaling_instance_group` [resource](../resources/autoscaling_instance_group.md).
