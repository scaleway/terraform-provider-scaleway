### With additional volumes and tags

resource "scaleway_block_volume" "data" {
  size_in_gb = 100
  iops       = 5000
}

resource "scaleway_instance_server" "web" {
  type  = "DEV1-S"
  image = "ubuntu_jammy"

  tags = ["hello", "public"]

  root_volume {
    delete_on_termination = false
  }

  additional_volume_ids = [scaleway_block_volume.data.id]
}
