resource "scaleway_vpc_private_network" "pn" {}

resource "scaleway_k8s_cluster" "cluster" {
  name                        = "tf-cluster"
  description                 = "cluster made in terraform"
  version                     = "1.32.3"
  cni                         = "calico"
  tags                        = ["terraform"]
  private_network_id          = scaleway_vpc_private_network.pn.id
  delete_additional_resources = false

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

resource "scaleway_k8s_pool" "pool" {
  cluster_id  = scaleway_k8s_cluster.cluster.id
  name        = "tf-pool"
  node_type   = "DEV1-M"
  size        = 3
  autoscaling = true
  autohealing = true
  min_size    = 1
  max_size    = 5
}