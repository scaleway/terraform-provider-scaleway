---
page_title: "Scaleway: scaleway_k8s_cluster"
description: |-
  Manages Scaleway Kubernetes clusters.
---

# scaleway_k8s_cluster

Creates and manages Scaleway Kubernetes clusters. For more information, see [the documentation](https://developers.scaleway.com/en/products/k8s/api/).

## Examples

### Basic

```hcl
resource "scaleway_k8s_cluster" "jack" {
  name    = "jack"
  version = "1.19.4"
  cni     = "cilium"
}

resource "scaleway_k8s_pool" "john" {
  cluster_id = scaleway_k8s_cluster.jack.id
  name       = "john"
  node_type  = "DEV1-M"
  size       = 1
}
```

### With additional configuration

```hcl
resource "scaleway_k8s_cluster" "john" {
  name             = "john"
  description      = "my awesome cluster"
  version          = "1.18.0"
  cni              = "calico"
  enable_dashboard = true
  ingress          = "traefik"
  tags             = ["i'm an awsome tag", "yay"]

  autoscaler_config {
    disable_scale_down              = false
    scale_down_delay_after_add      = "5m"
    estimator                       = "binpacking"
    expander                        = "random"
    ignore_daemonsets_utilization   = true
    balance_similar_node_groups     = true
    expendable_pods_priority_cutoff = -5
  }
}

resource "scaleway_k8s_pool" "john" {
  cluster_id  = scaleway_k8s_cluster.john.id
  name        = "john"
  node_type   = "DEV1-M"
  size        = 3
  autoscaling = true
  autohealing = true
  min_size    = 1
  max_size    = 5
}
```

### With the kubernetes provider

```hcl
resource "scaleway_k8s_cluster" "joy" {
  name    = "joy"
  version = "1.18.0"
  cni     = "flannel"
}

resource "scaleway_k8s_pool" "john" {
  cluster_id = scaleway_k8s_cluster.joy.id
  name       = "john"
  node_type  = "DEV1-M"
  size       = 1
}

resource "null_resource" "kubeconfig" {
  depends_on = [scaleway_k8s_pool.john] # at least one pool here
  triggers = {
    host                   = scaleway_k8s_cluster.joy.kubeconfig[0].host
    token                  = scaleway_k8s_cluster.joy.kubeconfig[0].token
    cluster_ca_certificate = scaleway_k8s_cluster.joy.kubeconfig[0].cluster_ca_certificate
  }
}

provider "kubernetes" {
  load_config_file = "false"

  host  = null_resource.kubeconfig.triggers.host
  token = null_resource.kubeconfig.triggers.token
  cluster_ca_certificate = base64decode(
  null_resource.kubeconfig.triggers.cluster_ca_certificate
  )
}
```

The `null_resource` is needed because when the cluster is created, it's status is `pool_required`, but the kubeconfig can already be downloaded.
It leads the `kubernetes` provider to start creating its objects, but the DNS entry for the Kubernetes master is not yet ready, that's why it's needed to wait for at least a pool.

## Arguments Reference

The following arguments are supported:

- `name` - (Required) The name for the Kubernetes cluster.

- `description` - (Optional) A description for the Kubernetes cluster.

- `version` - (Required) The version of the Kubernetes cluster.

- `cni` - (Required) The Container Network Interface (CNI) for the Kubernetes cluster.
~> **Important:** Updates to this field will recreate a new resource.

- `enable_dashboard` - (Defaults to `false`) Enables the [Kubernetes dashboard](https://github.com/kubernetes/dashboard) for the Kubernetes cluster.

- `ingress` - (Defaults to `none`) The [ingress controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) to be deployed on the Kubernetes cluster.

- `tags` - (Optional) The tags associated with the Kubernetes cluster.

- `autoscaler_config` - (Optional) The configuration options for the [Kubernetes cluster autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler).

    - `disable_scale_down` - (Defaults to `false`) Disables the scale down feature of the autoscaler.

    - `scale_down_delay_after_add` - (Defaults to `10m`) How long after scale up that scale down evaluation resumes.

    - `scale_down_unneeded_time` - (Default to `10m`) How long a node should be unneeded before it is eligible for scale down.

    - `estimator` - (Defaults to `binpacking`) Type of resource estimator to be used in scale up.

    - `expander` - (Default to `random`) Type of node group expander to be used in scale up.

    - `ignore_daemonsets_utilization` - (Defaults to `false`) Ignore DaemonSet pods when calculating resource utilization for scaling down.

    - `balance_similar_node_groups` - (Defaults to `false`) Detect similar node groups and balance the number of nodes between them.

    - `expendable_pods_priority_cutoff` - (Defaults to `-10`) Pods with priority below cutoff will be expendable. They can be killed without any consideration during scale down and they don't cause scale up. Pods with null priority (PodPriority disabled) are non expendable.

- `auto_upgrade` - (Optional) The auto upgrade configuration.

    - `enable` - (Optional) Set to `true` to enable Kubernetes patch version auto upgrades.
~> **Important:** When enabling auto upgrades, the `version` field take a minor version like x.y (ie 1.18).

    - `maintenance_window_start_hour` - (Optional) The start hour (UTC) of the 2-hour auto upgrade maintenance window (0 to 23).

    - `maintenance_window_day` - (Optional) The day of the auto upgrade maintenance window (`monday` to `sunday`, or `any`).

- `feature_gates` - (Optional) The list of [feature gates](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/) to enable on the cluster.

- `admission_plugins` - (Optional) The list of [admission plugins](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/) to enable on the cluster.

- `delete_additional_resources` - (Defaults to `false`) Delete additional resources like block volumes and loadbalancers that were created in Kubernetes on cluster deletion.

- `default_pool` - (Deprecated) See below.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the cluster should be created.

- `organization_id` - (Defaults to [provider](../index.md#organization_id) `organization_id`) The ID of the organization the cluster is associated with.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the cluster is associated with.


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
- `upgrade_available` - Set to `true` if a newer Kubernetes version is available.

## Import

Kubernetes clusters can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_k8s_cluster.mycluster fr-par/11111111-1111-1111-1111-111111111111
```

## Deprecation of default_pool

`default_pool` is deprecated in favour the `scaleway_k8s_pool` resource. Here is a migration example.

Before:

```hcl
resource "scaleway_k8s_cluster" "jack" {
  name    = "jack"
  version = "1.18.0"
  cni     = "cilium"

  default_pool {
    node_type = "DEV1-M"
    size      = 1
  }
}
```

After:

```hcl
resource "scaleway_k8s_cluster" "jack" {
  name    = "jack"
  version = "1.18.0"
  cni     = "cilium"
}

resource "scaleway_k8s_pool" "default" {
  cluster_id = scaleway_k8s_cluster.jack.id
  name       = "default"
  node_type  = "DEV1-M"
  size       = 1
}
```

Once you have moved all the `default_pool` into their own object, you will need to import them. If your pool had the ID 11111111-1111-1111-1111-111111111111 in the `fr-par` region, you can import it by typing:

```bash
$ terraform import scaleway_k8s_pool.default fr-par/11111111-1111-1111-1111-111111111111
```

Then you will only need to type `terraform apply` to have a smooth migration.
