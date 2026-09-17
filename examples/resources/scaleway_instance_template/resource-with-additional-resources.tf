# Security Group
resource "scaleway_instance_security_group" "sg" {}
# Placement Group
resource "scaleway_instance_placement_group" "pg" {}
# VPC + Private Network
resource "scaleway_vpc" "vpc" {}
resource "scaleway_vpc_private_network" "pn" {
  vpc_id = scaleway_vpc.vpc.id
}
# Filesystem
resource "scaleway_file_filesystem" "fs" {
  size_in_gb = 25
}

resource "scaleway_instance_template" "main" {
  name        = "instance-template-additional-resources"
  server_type = "PRO2-M"

  security_group_id  = scaleway_instance_security_group.sg.id
  placement_group_id = scaleway_instance_placement_group.pg.id
  private_networks   = [scaleway_vpc_private_network.pn.id]
  filesystem_ids     = [scaleway_file_filesystem.fs.id]
}
