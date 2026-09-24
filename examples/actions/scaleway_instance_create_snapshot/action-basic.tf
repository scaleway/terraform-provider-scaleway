### Create a snapshot from an instance volume

resource "scaleway_instance_server" "main" {
  name          = "my-instance"
  image         = "ubuntu_jammy"
  type          = "DEV1-S"
  enable_ipv6   = false
  is_blocked    = false
  state         = "running"
  wait_for_boot = true

  root_volume {
    size = 20
  }
}

action "scaleway_instance_create_snapshot" "create_snapshot" {
  config {
    volume_id = scaleway_instance_server.main.root_volume.0.volume_id
    name      = "my-instance-snapshot"
    wait      = true
  }
}
