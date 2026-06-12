resource "scaleway_vpc" "main" {
  name = "my-vpc"
}

resource "scaleway_vpc_private_network" "pn" {
  name   = "my-private-network"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_opensearch_deployment" "main" {
  name        = "my-opensearch-cluster"
  version     = "2.0"
  node_amount = 1
  node_type   = "SEARCHDB-DEDICATED-2C-8G"
  password    = "ThisIsASecurePassword123!"

  private_network {
    private_network_id = scaleway_vpc_private_network.pn.id
  }

  volume {
    type       = "sbs_5k"
    size_in_gb = 5
  }
}
