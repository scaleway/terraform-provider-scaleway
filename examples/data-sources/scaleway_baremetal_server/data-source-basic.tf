## Basic

# Get info by server name
data "scaleway_baremetal_server" "by_name" {
  name = "foobar"
  zone = "fr-par-2"
}

# Get info by server id
data "scaleway_baremetal_server" "by_id" {
  server_id = "11111111-1111-1111-1111-111111111111"
}
