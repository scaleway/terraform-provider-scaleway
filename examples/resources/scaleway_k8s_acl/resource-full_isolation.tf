resource "scaleway_vpc_private_network" "acl_basic" {}

resource "scaleway_k8s_cluster" "acl_basic" {
  name                        = "acl-basic"
  version                     = "1.32.2"
  cni                         = "cilium"
  delete_additional_resources = true
  private_network_id          = scaleway_vpc_private_network.acl_basic.id
}

resource "scaleway_k8s_acl" "acl_basic" {
  cluster_id    = scaleway_k8s_cluster.acl_basic.id
  no_ip_allowed = true
}
