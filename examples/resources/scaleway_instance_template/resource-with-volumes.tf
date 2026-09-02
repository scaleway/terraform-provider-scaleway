resource "scaleway_block_volume" "sbs" {
  size_in_gb = 15
  iops       = 5000
}
resource "scaleway_block_snapshot" "snap" {
  volume_id = scaleway_block_volume.sbs.id
}

resource "scaleway_instance_template" "main" {
  name        = "instance-template-with-volumes"
  server_type = "GP1-L"

  volumes = [
    {
      volume_type = "l_ssd"
      image_label = "debian_trixie"
      size_in_gb  = 25
      tags        = ["local"]
      name        = "root-volume"
    },
    {
      volume_type      = "sbs"
      base_snapshot_id = scaleway_block_snapshot.snap.id
      size_in_gb       = 30
      perf_iops        = 15000
    }
  ]
}
