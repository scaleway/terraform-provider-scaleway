# Get info by volume name
data "scaleway_instance_volume" "my_volume" {
  name = "my-volume-name"
}

# Get info by volume ID
data "scaleway_instance_volume" "my_volume" {
  volume_id = "11111111-1111-1111-1111-111111111111"
}
