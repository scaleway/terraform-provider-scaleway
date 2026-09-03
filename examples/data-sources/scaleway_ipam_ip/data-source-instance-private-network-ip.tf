### Instance Private Network IP

# Connect your instance to a private network using a private nic.
resource "scaleway_instance_private_nic" "nic" {
  server_id          = scaleway_instance_server.server.id
  private_network_id = scaleway_vpc_private_network.pn.id
}

# Find server private IPv4 using private-nic mac address
data "scaleway_ipam_ip" "by_mac" {
  mac_address = scaleway_instance_private_nic.nic.mac_address
  type        = "ipv4"
}

# Find server private IPv4 using private-nic id
data "scaleway_ipam_ip" "by_id" {
  resource {
    id   = scaleway_instance_private_nic.nic.id
    type = "instance_private_nic"
  }
  type = "ipv4"
}
