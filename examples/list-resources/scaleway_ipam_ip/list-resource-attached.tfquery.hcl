# List only IPAM IPs that are attached to a resource
list "scaleway_ipam_ip" "attached" {
  provider = scaleway

  config {
    regions  = ["fr-par"]
    attached = true
  }
}
