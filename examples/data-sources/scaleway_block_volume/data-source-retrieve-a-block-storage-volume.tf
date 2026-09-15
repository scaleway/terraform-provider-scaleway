## Retrieve a Block Storage volume

// Get info by volume name
data "scaleway_block_volume" "my_volume" {
  name = "my-name"
}

// Get info by volume ID
data "scaleway_block_volume" "my_volume" {
  volume_id = "11111111-1111-1111-1111-111111111111"
}
