### Basic

resource "scaleway_instance_private_nic" "pnic01" {
  server_id          = "fr-par-1/11111111-1111-1111-1111-111111111111"
  private_network_id = "fr-par-1/aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
}
