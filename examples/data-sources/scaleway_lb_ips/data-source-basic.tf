### Basic

# Find multiple IPs that share the same CIDR block
data "scaleway_lb_ips" "my_key" {
  ip_cidr_range = "0.0.0.0/0"
}
# Find IPs by CIDR block and zone
data "scaleway_lb_ips" "my_key" {
  ip_cidr_range = "0.0.0.0/0"
  zone          = "fr-par-2"
}

# Find IPs that share the same tags and type
data "scaleway_lb_ips" "ips_by_tags_and_type" {
  tags    = ["a tag"]
  ip_type = "ipv4"
}
