# List IPAM IPs filtered by tag
list "scaleway_ipam_ip" "by_tag" {
  provider = scaleway

  config {
    regions = ["*"]
    tags    = ["production"]
  }
}
