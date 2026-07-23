## Retrieve a volume's snapshot

// Get info by snapshot name
data "scaleway_block_snapshot" "my_snapshot" {
  name = "my-name"
}

// Get info by snapshot name and volume id
data "scaleway_block_snapshot" "my_snapshot" {
  name      = "my-name"
  volume_id = "11111111-1111-1111-1111-111111111111"
}

// Get info by snapshot ID
data "scaleway_block_snapshot" "my_snapshot" {
  snapshot_id = "11111111-1111-1111-1111-111111111111"
}
