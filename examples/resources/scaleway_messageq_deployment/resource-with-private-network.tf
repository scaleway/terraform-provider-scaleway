resource "scaleway_vpc" "main" {
  name = "my-vpc"
}

resource "scaleway_vpc_private_network" "pn" {
  name   = "my-private-network"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_messageq_deployment" "main" {
  name       = "my-messageq"
  version    = "4.0"
  node_count = 1
  node_type  = "MESSAGEQ-DEDICATED-2C-8G"
  user_name  = "admin"
  password   = "ThisIsASecurePassword123!"

  private_network {
    private_network_id = scaleway_vpc_private_network.pn.id
  }

  volume {
    type       = "sbs_5k"
    size_in_gb = 5
  }
}
