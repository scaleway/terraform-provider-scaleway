### Basic

resource "scaleway_vpc" "vpc01" {
  name = "tf-vpc-acl"
}

resource "scaleway_vpc_acl" "acl01" {
  vpc_id  = scaleway_vpc.vpc01.id
  is_ipv6 = false
  rules {
    protocol      = "TCP"
    src_port_low  = 0
    src_port_high = 0
    dst_port_low  = 80
    dst_port_high = 80
    source        = "0.0.0.0/0"
    destination   = "0.0.0.0/0"
    description   = "Allow HTTP traffic from any source"
    action        = "accept"
  }
  default_policy = "drop"
}
