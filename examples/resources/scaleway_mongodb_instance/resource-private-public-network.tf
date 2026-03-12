### MongoDB instance with Private Network and Public Network

resource "scaleway_vpc_private_network" "pn01" {
  name   = "my_private_network"
  region = "fr-par"
}

resource "scaleway_mongodb_instance" "main" {
  name              = "test-mongodb-basic1"
  version           = "7.0.12"
  node_type         = "MGDB-PLAY2-NANO"
  node_number       = 1
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  volume_size_in_gb = 5

  private_network {
    pn_id = scaleway_vpc_private_network.pn01.id
  }

  public_network {}
}
