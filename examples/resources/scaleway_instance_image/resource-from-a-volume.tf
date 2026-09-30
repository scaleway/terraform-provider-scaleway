### From a volume

resource "scaleway_instance_volume" "volume" {
  type       = "b_ssd"
  size_in_gb = 20
}

resource "scaleway_instance_snapshot" "volume_snapshot" {
  volume_id = scaleway_instance_volume.volume.id
}

resource "scaleway_instance_image" "volume_image" {
  name           = "image_from_volume"
  root_volume_id = scaleway_instance_snapshot.volume_snapshot.id
}
