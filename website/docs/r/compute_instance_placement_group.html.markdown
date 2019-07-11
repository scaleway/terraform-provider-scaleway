---
layout: "scaleway"
page_title: "Scaleway: scaleway_compute_instance_placement_group"
description: |-
  Manages Scaleway Compute Instance Placement Groups (aka. Compute Clusters).
---

# scaleway_compute_instance_placement_group

Creates and manages Compute Instance Placement Groups (aka. Compute Clusters). For more information, see [the documentation](https://developers.scaleway.com/en/products/instance/api/#compute-clusters-7fd7e0).

## Example Usage

```hcl
resource "scaleway_compute_instance_placement_group" "availability_group" {}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Optional) The name of the placement group.
- `policy_type` - (Defaults to `low_latency`) The [policy type](https://developers.scaleway.com/en/products/instance/api/#compute-clusters-7fd7e0) of the placement group. Possible values are: `low_latency` or `max_availability`.
- `policy_mode` - (Defaults to `optional`) The [policy mode](https://developers.scaleway.com/en/products/instance/api/#compute-clusters-7fd7e0) of the placement group. Possible values are: `optional` or `enforced`.
- `zone` - (Defaults to [provider](../index.html#zone) `zone`) The [zone](../guides/regions_and_zones.html#zones) in which the placement group should be created.
- `project_id` - (Defaults to [provider](../index.html#project_id) `project_id`) The ID of the project the placement group is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the placement group.
- `policy_respected` - Is true when the policy is respected.

## Import

Placement groups can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_compute_instance_placement_group.availability_group fr-par-1/11111111-1111-1111-1111-111111111111
```
