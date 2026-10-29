# Get the IPv6 ACL for a VPC
data "scaleway_vpc_acl" "my_acl_v6" {
  vpc_id  = scaleway_vpc.my_vpc.id
  is_ipv6 = true
}
