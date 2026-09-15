### By type and resource

resource "scaleway_vpc" "vpc01" {
  name = "my vpc"
}

resource "scaleway_vpc_private_network" "pn01" {
  vpc_id = scaleway_vpc.vpc01.id
  ipv4_subnet {
    subnet = "172.16.32.0/22"
  }
}

resource "scaleway_redis_cluster" "redis01" {
  name         = "my_redis_cluster"
  version      = "7.0.5"
  node_type    = "RED1-XS"
  user_name    = "my_initial_user"
  password     = "thiZ_is_v&ry_s3cret"
  cluster_size = 3
  private_network {
    id = scaleway_vpc_private_network.pn01.id
  }
}

data "scaleway_ipam_ips" "by_type_and_resource" {
  type = "ipv4"
  resource {
    id   = scaleway_redis_cluster.redis01.id
    type = "redis_cluster"
  }
}
