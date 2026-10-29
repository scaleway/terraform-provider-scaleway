# Get info by pool name (need cluster_id)
data "scaleway_k8s_pool" "my_key" {
  name       = "my-pool-name"
  cluster_id = "11111111-1111-1111-1111-111111111111"
}
