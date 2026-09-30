data "scaleway_instance_private_nic" "by_nic_id" {
  server_id      = "11111111-1111-1111-1111-111111111111"
  private_nic_id = "11111111-1111-1111-1111-111111111111"
}

data "scaleway_instance_private_nic" "by_pn_id" {
  server_id          = "11111111-1111-1111-1111-111111111111"
  private_network_id = "11111111-1111-1111-1111-111111111111"
}

data "scaleway_instance_private_nic" "by_tags" {
  server_id = "11111111-1111-1111-1111-111111111111"
  tags      = ["mytag"]
}
