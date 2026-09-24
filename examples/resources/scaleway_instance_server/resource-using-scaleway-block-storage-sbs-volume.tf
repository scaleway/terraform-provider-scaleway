#### Using Scaleway Block Storage (SBS) volume

resource "scaleway_instance_server" "server" {
  type  = "PLAY2-MICRO"
  image = "ubuntu_jammy"
  root_volume {
    volume_type = "sbs_volume"
    sbs_iops    = 15000
    size_in_gb  = 50
  }
}
