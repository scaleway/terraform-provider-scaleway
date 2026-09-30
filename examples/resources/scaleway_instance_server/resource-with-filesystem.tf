### With filesystem

resource "scaleway_block_volume" "volume" {
  iops       = 15000
  size_in_gb = 15
}

resource "scaleway_file_filesystem" "terraform_instance_filesystem" {
  name       = "filesystem-instance-terraform"
  size_in_gb = 100
}

resource "scaleway_instance_server" "base" {
  type  = "POP2-HM-2C-16G"
  state = "started"
  tags  = ["terraform-test", "scaleway_instance_server", "state"]
  root_volume {
    volume_type = "sbs_volume"
    volume_id   = scaleway_block_volume.volume.id
  }
  filesystems {
    filesystem_id = scaleway_file_filesystem.terraform_instance_filesystem.id
  }
}
