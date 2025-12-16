# Get info by cluster id
data "scaleway_k8s_cluster" "my_key" {
  cluster_id = "11111111-1111-1111-1111-111111111111"
}