---
layout: "scaleway"
page_title: "Scaleway: scaleway_k8s_pool_beta"
description: |-
  Manages Scaleway Kubernetes cluster pools.
---

# scaleway_k8s_pool_beta

Creates and manages Scaleway Kubernetes cluster pools. For more information, see [the documentation](https://developers.scaleway.com/en/products/k8s/api/).

## Examples

### Basic

```hcl
resource "scaleway_k8s_cluster_beta" "jack" {
  name = "jack"
  version = "1.16.1"
  cni = "calico"
  default_pool {
    node_type = "gp1_xs"
    size = 3
  }
}

resource "scaleway_k8s_pool_beta" "bill" {
  cluster_id = "${scaleway_k8s_cluster_beta.jack.id}"
  name = "bill"
  node_type = "gp1_s"
  size = 3
  min_size = 0
  max_size = 10
  autoscaling = true
  autohealing = true
  container_runtime = "docker"
  placement_group_id = "1267e3fd-a51c-49ed-ad12-857092ee3a3d"
}
```

## Arguments Reference

The following arguments are supported:

- `cluster_id` - (Required) The ID of the Kubernetes cluster on which this pool will be created.

- `name` - (Required) The name for the pool.
~> **Important:** Updates to this field will recreate a new resource.

- `node_type` - (Required)  The commercial type of the pool instances.
~> **Important:** Updates to this field will recreate a new resource.

- `size` - (Required) The size of the pool.

- `min_size` - (Defaults to `1`) The minimum size of the pool, used by the autoscaling feature.

- `max_size` - (Defaults to `size`) The maximum size of the pool, used by the autoscaling feature.

- `placement_group_id` - (Optional) The [placement group](https://developers.scaleway.com/en/products/instance/api/#compute-clusters-7fd7e0) the nodes of the pool will be attached to.

- `autoscaling` - (Defaults to `false`) Enables the autoscaling feature for this pool.
~> **Important:** When enabled, an update of the `size` will not be taken into account.

- `autohealing` - (Defaults to `false`) Enables the autohealing feature for this pool.

- `container_runtime` - (Defaults to `docker`) The container runtime of the pool.

- `region` - (Defaults to [provider](../index.html#region) `region`) The [region](../guides/regions_and_zones.html#regions) in which the pool should be created.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the pool.
- `created_at` - The creation date of the pool.
- `updated_at` - The last update date of the pool.
- `version` - The version of the pool.

## Import

Kubernetes pools can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_k8s_pool_beta.mypool fr-par/11111111-1111-1111-1111-111111111111
```
