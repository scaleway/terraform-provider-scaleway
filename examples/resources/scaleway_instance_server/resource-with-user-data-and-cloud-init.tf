### With user data and cloud-init

resource "scaleway_instance_server" "web" {
  type  = "DEV1-S"
  image = "ubuntu_jammy"

  user_data = {
    foo        = "bar"
    cloud-init = file("${path.module}/cloud-init.yml")
  }
}
