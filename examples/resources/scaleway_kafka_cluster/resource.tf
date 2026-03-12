resource "scaleway_vpc" "main" {
  region = "fr-par"
  name   = "my-vpc"
}

resource "scaleway_vpc_private_network" "pn" {
  name   = "my-private-network"
  region = "fr-par"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_kafka_cluster" "main" {
  name              = "my-kafka-cluster"
  version           = "3.9.0"
  node_amount       = 1
  node_type         = "KAFK-PLAY-NANO"
  volume_type       = "sbs_5k"
  volume_size_in_gb = 10
  user_name         = "admin"
  password          = "thiZ_is_v&ry_s3cret"

  private_network {
    pn_id = scaleway_vpc_private_network.pn.id
  }
}
