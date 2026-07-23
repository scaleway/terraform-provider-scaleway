### From a server

resource "scaleway_instance_server" "server" {
  image = "ubuntu_jammy"
  type  = "DEV1-S"
}

resource "scaleway_instance_snapshot" "server_snapshot" {
  volume_id = scaleway_instance_server.main.root_volume.0.volume_id
}

resource "scaleway_instance_image" "server_image" {
  name           = "image_from_server"
  root_volume_id = scaleway_instance_snapshot.server_snapshot.id
}
