resource "scaleway_vpc" "main" {}

resource "scaleway_vpc_private_network" "main" {
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_k8s_cluster" "main" {
  version                     = "1.35.3"
  cni                         = "cilium"
  name                        = "example-cluster"
  delete_additional_resources = true
  private_network_id          = scaleway_vpc_private_network.main.id
}

resource "scaleway_instance_placement_group" "main" {}

resource "scaleway_k8s_pool" "main" {
  cluster_id         = scaleway_k8s_cluster.main.id
  version            = scaleway_k8s_cluster.main.version
  node_type          = "DEV1-M"
  size               = 3
  placement_group_id = scaleway_instance_placement_group.main.id

  # Make sure that the new resource is created before destroying the old one on changes that require replacement
  lifecycle {
    create_before_destroy = true
  }
}
