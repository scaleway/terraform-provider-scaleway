# Get info by server name
data "scaleway_instance_server" "my_key" {
  name = "my-server-name"
}

# Get info by server id
data "scaleway_instance_server" "my_key" {
  server_id = "11111111-1111-1111-1111-111111111111"
}
