### Redis cluster with a Private Network

resource "scaleway_vpc_private_network" "pn" {
  name = "private-network"
}

resource "scaleway_redis_cluster" "main" {
  name         = "test_redis_endpoints"
  version      = "6.2.7"
  node_type    = "RED1-MICRO"
  user_name    = "my_initial_user"
  password     = "thiZ_is_v&ry_s3cret"
  cluster_size = 1
  private_network {
    id = scaleway_vpc_private_network.pn.id
    service_ips = [
      "10.12.1.1/20",
    ]
  }
  depends_on = [
    scaleway_vpc_private_network.pn
  ]
}
