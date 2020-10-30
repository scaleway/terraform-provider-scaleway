---
page_title: "Scaleway: scaleway_instance_placement_group"
description: |-
  Manages Scaleway Compute Instance Placement Groups.
---

# scaleway_instance_placement_group

Creates and manages Compute Instance Placement Groups. For more information, see [the documentation](https://developers.scaleway.com/en/products/instance/api/#placement-groups-d8f653).

## Example Usage

```hcl
resource "scaleway_instance_placement_group" "availability_group" {}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Optional) The name of the placement group.
- `policy_type` - (Defaults to `max_availability`) The [policy type](https://developers.scaleway.com/en/products/instance/api/#placement-groups-d8f653) of the placement group. Possible values are: `low_latency` or `max_availability`.
- `policy_mode` - (Defaults to `optional`) The [policy mode](https://developers.scaleway.com/en/products/instance/api/#placement-groups-d8f653) of the placement group. Possible values are: `optional` or `enforced`.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the placement group should be created.
- `organization_id` - (Defaults to [provider](../index.md#organization_id) `organization_id`) The ID of the organization the placement group is associated with.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the placement group is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the placement group.
- `policy_respected` - Is true when the policy is respected.

## Import

Placement groups can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_instance_placement_group.availability_group fr-par-1/11111111-1111-1111-1111-111111111111
```
