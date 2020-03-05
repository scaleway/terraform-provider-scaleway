---
layout: "scaleway"
page_title: "Scaleway: scaleway_k8s_cluster_beta"
description: |-
  Manages Scaleway Kubernetes clusters.
---

# scaleway_k8s_cluster_beta

Creates and manages Scaleway Kubernetes clusters. For more information, see [the documentation](https://developers.scaleway.com/en/products/k8s/api/).

## Examples

### Basic

```hcl
resource "scaleway_k8s_cluster_beta" "jack" {
  name = "jack"
  version = "1.16.1"
  cni = "calico"
  default_pool {
    node_type = "GP1-XS"
    size = 3
  }
}
```

### With additional configuration

```hcl
resource "scaleway_k8s_cluster_beta" "john" {
  name = "john"
  description = "my awesome cluster"
  version = "1.16.1"
  cni = "weave"
  enable_dashboard = true
  ingress = "traefik"
  tags = ["i'm an awsome tag", "yay"]

  default_pool {
    node_type = "GP1-XS"
    size = 3
    autoscaling = true
    autohealing = true
    min_size = 1
    max_size = 5
  }

  autoscaler_config {
    disable_scale_down = false
    scale_down_delay_after_add = 5m
    estimator = "binpacking"
    expander = "random"
    ignore_daemonsets_utilization = true
    balance_similar_node_groups = true
    expendable_pods_priority_cutoff = -5
  }
}
```

### With the kubernetes provider

```hcl
resource "scaleway_k8s_cluster_beta" "joy" {
  name = "joy"
  version = "1.16.1"
  cni = "flannel"
  default_pool {
    node_type = "GP1-XS"
    size = 3
  }
}

provider "kubernetes" {
  host  = scaleway_k8s_cluster_beta.joy.kubeconfig[0].host
  token  = scaleway_k8s_cluster_beta.joy.kubeconfig[0].token
  cluster_ca_certificate = base64decode(
    scaleway_k8s_cluster_beta.joy.kubeconfig[0].cluster_ca_certificate
  )
}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Required) The name for the Kubernetes cluster.

- `description` - (Optional) A description for the Kubernetes cluster.

- `version` - (Required) The version of the Kubernetes cluster.

- `cni` - (Required) The Container Network Interface (CNI) for the Kubernetes cluster.
~> **Important:** Updates to this field will recreate a new resource.

- `enable_dashboard` - (Defaults to `false`) Enables the [Kubernetes dashboard](https://github.com/kubernetes/dashboard) for the Kubernetes cluster.
~> **Important:** Updates to this field will recreate a new resource.

- `ingress` - (Defaults to `none`) The [ingress controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) to be deployed on the Kubernetes cluster.
~> **Important:** Updates to this field will recreate a new resource.

- `tags` - (Optional) The tags associated with the Kubernetes cluster.

- `autoscaler_config` - (Optional) The configuration options for the [Kubernetes cluster autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler).

  - `disable_scale_down` - (Defaults to `false`) Disables the scale down feature of the autoscaler.

  - `scale_down_delay_after_add` - (Defaults to `10m`) How long after scale up that scale down evaluation resumes.

  - `estimator` - (Defaults to `binpacking`) Type of resource estimator to be used in scale up.

  - `expander` - (Default to `random`) Type of node group expander to be used in scale up.

  - `ignore_daemonsets_utilization` - (Defaults to `false`) Ignore DaemonSet pods when calculating resource utilization for scaling down.

  - `balance_similar_node_groups` - (Defaults to `false`) Detect similar node groups and balance the number of nodes between them.

  - `expendable_pods_priority_cutoff` - (Defaults to `-10`) Pods with priority below cutoff will be expendable. They can be killed without any consideration during scale down and they don't cause scale up. Pods with null priority (PodPriority disabled) are non expendable.

- `auto_upgrade` - (Optional) The auto upgrade configuration.

  - `enable` - (Optional) Set to `true` to enable Kubernetes patch version auto upgrades.

  - `maintenance_window_start_hour` - (Optional) The start hour (UTC) of the 2-hour auto upgrade maintenance window (0 to 23).

  - `maintenance_window_day` - (Optional) The day of the auto upgrade maintenance window (`monday` to `sunday`, or `any`).

- `feature_gates` - (Optional) The list of [feature gates](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/) to enable on the cluster.

- `admission_controllers` - (Optional) The list of [admission plugins](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/) to enable on the cluster.

- `default_pool` - (Required) The cluster's default pool configuration.
  
  - `node_type` - (Required)  The commercial type of the default pool instances.
~> **Important:** Updates to this field will recreate a new default pool in a rolling upgrade fashion. It will first create the new pool, wait until all the nodes are ready and then delete the old pool. Errors may occur if you don't have enough quotas.

  - `size` - (Required) The size of the default pool.

  - `min_size` - (Defaults to `1`) The minimum size of the default pool, used by the autoscaling feature.

  - `max_size` - (Defaults to `size`) The maximum size of the default pool, used by the autoscaling feature.

  - `placement_group_id` - (Optional) The [placement group](https://developers.scaleway.com/en/products/instance/api/#placement-groups-d8f653) the nodes of the pool will be attached to.
~> **Important:** Updates to this field will recreate a new default pool.

  - `autoscaling` - (Defaults to `false`) Enables the autoscaling feature for the default pool.
~> **Important:** When enabled, an update of the `size` will not be taken into account.

  - `autohealing` - (Defaults to `false`) Enables the autohealing feature for the default pool.

  - `container_runtime` - (Defaults to `docker`) The container runtime of the default pool.
~> **Important:** Updates to this field will recreate a new default pool.

  - `wait_for_pool_ready` - (Default to `false`) Whether to wait for the pool to be ready.

- `region` - (Defaults to [provider](../index.html#region) `region`) The [region](../guides/regions_and_zones.html#regions) in which the cluster should be created.

- `organization_id` - (Defaults to [provider](../index.html#organization_id) `organization_id`) The ID of the organization the cluster is associated with.


## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the cluster.
- `created_at` - The creation date of the cluster.
- `updated_at` - The last update date of the cluster.
- `apiserver_url` - The URL of the Kubernetes API server.
- `wildcard_dns` - The DNS wildcard that points to all ready nodes.
- `kubeconfig`
  - `config_file` - The raw kubeconfig file.
  - `host` - The URL of the Kubernetes API server.
  - `cluster_ca_certificate` - The CA certificate of the Kubernetes API server.
  - `token` - The token to connect to the Kubernetes API server.
- `status` - The status of the Kubernetes cluster.
- `default_pool`
  - `pool_id` - The ID of the default pool.
  - `status` - The status of the default pool.
  - `nodes` - (List of) The nodes in the default pool. Defined below.
  - `created_at` - The creation date of the default pool.
  - `updated_at` - The last update date of the default pool.
- `upgrade_available` - Set to `true` if a newer Kubernetes version is available.

## nodes

- `name` - The name of the node.
- `public_ip` - The public IPv4.
- `public_ip_v6` - The public IPv6.
- `status` - The status of the node.


## Import

Kubernetes clusters can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_k8s_cluster_beta.mycluster fr-par/11111111-1111-1111-1111-111111111111
```
