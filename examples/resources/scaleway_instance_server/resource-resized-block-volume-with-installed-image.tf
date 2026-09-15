#### Resized block volume with installed image

resource "scaleway_instance_server" "image" {
  type  = "PRO2-XXS"
  image = "ubuntu_jammy"
  root_volume {
    size_in_gb = 100
  }
}
