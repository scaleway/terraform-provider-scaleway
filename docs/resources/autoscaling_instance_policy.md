---
subcategory: "Autoscaling"
page_title: "Scaleway: scaleway_autoscaling_instance_policy"
---

# Resource: scaleway_autoscaling_instance_policy

Books and manages Autoscaling Instance policies.

## Example Usage

### Basic

```terraform
resource "scaleway_autoscaling_instance_policy" "up" {
  instance_group_id = scaleway_autoscaling_instance_group.main.id
  name              = "scale-up-if-cpu-high"
  action            = "scale_up"
  type              = "flat_count"
  value             = 1
  priority          = 1

  metric {
    name               = "cpu scale up"
    managed_metric     = "managed_metric_instance_cpu"
    operator           = "operator_greater_than"
    aggregate          = "aggregate_average"
    sampling_range_min = 5
    threshold          = 70
  }
}

resource "scaleway_autoscaling_instance_policy" "down" {
  instance_group_id = scaleway_autoscaling_instance_group.main.id
  name              = "scale-down-if-cpu-low"
  action            = "scale_down"
  type              = "flat_count"
  value             = 1
  priority          = 2

  metric {
    name               = "cpu scale down"
    managed_metric     = "managed_metric_instance_cpu"
    operator           = "operator_less_than"
    aggregate          = "aggregate_average"
    sampling_range_min = 5
    threshold          = 40
  }
```

## Argument Reference

The following arguments are supported:

- `instance_group_id` - (Required) The ID of the Instance group related to this policy.
- `name` - (Optional) The Instance policy name.
- `action` - (Required) The action to execute when the metric-based condition is met.
- `type` - (Required) How to use the number defined in `value` when determining by how many Instances to scale up/down.
- `value` - (Required) The value representing the magnitude of the scaling action to take for the Instance group. Depending on the `type` parameter, this number could represent a total number of Instances in the group, a number of Instances to add, or a percentage to scale the group by.
- `priority` - (Required) The priority of this policy compared to all other scaling policies. This determines the processing order. The lower the number, the higher the priority.
- `metric` - (Optional) Cockpit metric to use when determining whether to trigger a scale up/down action.
  - `name` - Name or description of the metric policy.
  - `operator` - Operator used when comparing the threshold value of the chosen `metric` to the actual sampled and aggregated value.
  - `aggregate` - How the values sampled for the `metric` should be aggregated.
  - `managed_metric` - The managed metric to use for this policy. These are available by default in Cockpit without any configuration or `node_exporter`. The chosen metric forms the basis of the condition that will be checked to determine whether a scaling action should be triggered.
  - `cockpit_metric_name` - The custom metric to use for this policy. This must be stored in Scaleway Cockpit. The metric forms the basis of the condition that will be checked to determine whether a scaling action should be triggered
  - `sampling_range_min` - The Interval of time, in minutes, during which metric is sampled.
  - `threshold` - The threshold value to measure the aggregated sampled `metric` value against. Combined with the `operator` field, determines whether a scaling action should be triggered.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the Instance policy exists.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the Project the Instance policy is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Instance policy.

~> **Important:** Autoscaling policies IDs are [zonal](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

## Import

Autoscaling instance policies can be imported using `{zone}/{id}`, e.g.

```bash
terraform import scaleway_autoscaling_instance_policy.main fr-par-1/11111111-1111-1111-1111-111111111111
```
