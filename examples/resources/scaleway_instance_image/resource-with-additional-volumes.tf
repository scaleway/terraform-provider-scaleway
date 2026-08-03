### With additional volumes

resource "scaleway_instance_server" "server" {
  image = "ubuntu_jammy"
  type  = "DEV1-S"
}

resource "scaleway_instance_volume" "volume" {
  type       = "b_ssd"
  size_in_gb = 20
}

resource "scaleway_instance_snapshot" "volume_snapshot" {
  volume_id = scaleway_instance_volume.volume.id
}
resource "scaleway_instance_snapshot" "server_snapshot" {
  volume_id = scaleway_instance_server.main.root_volume.0.volume_id
}

resource "scaleway_instance_image" "image" {
  name           = "image_with_extra_volumes"
  root_volume_id = scaleway_instance_snapshot.server_snapshot.id
  additional_volume_ids = [
    scaleway_instance_snapshot.volume_snapshot.id
  ]
}
