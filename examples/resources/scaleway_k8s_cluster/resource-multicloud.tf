# Multicloud Kubernetes Cluster Example
# For a detailed example of how to add or run Elastic Metal servers instead of Instances on your cluster, please refer to [this guide](../guides/multicloud_cluster_with_baremetal_servers.md).

resource "scaleway_k8s_cluster" "cluster" {
  name                        = "tf-cluster"
  type                        = "multicloud"
  version                     = "1.32.3"
  cni                         = "kilo"
  delete_additional_resources = false
}

resource "scaleway_k8s_pool" "pool" {
  cluster_id = scaleway_k8s_cluster.cluster.id
  name       = "tf-pool"
  node_type  = "external"
  size       = 0
  min_size   = 0
}
