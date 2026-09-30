### Create a Block Storage volume

resource "scaleway_block_volume" "block_volume" {
  iops       = 5000
  name       = "some-volume-name"
  size_in_gb = 20
}
