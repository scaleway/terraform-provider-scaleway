### With Private Network

resource "scaleway_vpc" "main" {
  name = "my-vpc"
}

resource "scaleway_vpc_private_network" "pn" {
  name   = "my-private-network"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_datawarehouse_deployment" "main" {
  name          = "my-datawarehouse"
  version       = "v25"
  replica_count = 1
  cpu_min       = 2
  cpu_max       = 4
  ram_per_cpu   = 4
  password      = "thiZ_is_v&ry_s3cret"

  private_network {
    pn_id = scaleway_vpc_private_network.pn.id
  }
}
