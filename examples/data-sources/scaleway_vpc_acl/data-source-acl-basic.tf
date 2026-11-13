# Get the IPv4 ACL for a VPC
data "scaleway_vpc_acl" "my_acl" {
  vpc_id  = scaleway_vpc.my_vpc.id
  is_ipv6 = false
}
