resource "scaleway_k8s_cluster" "main" {
  version = "1.32.3"
  cni     = "cilium"
}

resource "scaleway_k8s_pool" "main" {
  cluster_id         = scaleway_k8s_cluster.main.id
  node_type          = "DEV1-M"
  size               = 3
  min_size           = 0
  max_size           = 10
  autoscaling        = true
  autohealing        = true
  container_runtime  = "containerd"
  placement_group_id = "1267e3fd-a51c-49ed-ad12-857092ee3a3d"

  # Make sure that the new resource is created before destroying the old one on changes that require replacement
  lifecycle {
    create_before_destroy = true
  }
}
