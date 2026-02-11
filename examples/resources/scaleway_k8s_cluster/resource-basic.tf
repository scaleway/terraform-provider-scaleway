resource "scaleway_vpc_private_network" "pn" {}

resource "scaleway_k8s_cluster" "cluster" {
  name                        = "tf-cluster"
  version                     = "1.32.3"
  cni                         = "cilium"
  private_network_id          = scaleway_vpc_private_network.pn.id
  delete_additional_resources = false
}

resource "scaleway_k8s_pool" "pool" {
  cluster_id = scaleway_k8s_cluster.cluster.id
  name       = "tf-pool"
  node_type  = "DEV1-M"
  size       = 1
}
