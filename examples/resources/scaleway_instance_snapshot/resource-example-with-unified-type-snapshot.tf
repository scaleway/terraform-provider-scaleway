### Example with Unified type snapshot

resource "scaleway_instance_volume" "main" {
  type       = "l_ssd"
  size_in_gb = 10
}

resource "scaleway_instance_server" "main" {
  image = "ubuntu_jammy"
  type  = "DEV1-S"
  root_volume {
    size_in_gb  = 10
    volume_type = "l_ssd"
  }
  additional_volume_ids = [
    scaleway_instance_volume.main.id
  ]
}

resource "scaleway_instance_snapshot" "main" {
  volume_id  = scaleway_instance_volume.main.id
  depends_on = [scaleway_instance_server.main]
}
