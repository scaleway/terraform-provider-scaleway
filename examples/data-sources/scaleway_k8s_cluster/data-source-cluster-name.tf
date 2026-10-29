# Get info by cluster name
data "scaleway_k8s_cluster" "my_key" {
  name = "my-cluster-name"
}
