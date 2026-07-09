# Get info by IP address
data "scaleway_instance_ip" "my_ip" {
  address = "0.0.0.0"
}

# Get info by ID
data "scaleway_instance_ip" "my_ip" {
  id = "fr-par-1/11111111-1111-1111-1111-111111111111"
}
