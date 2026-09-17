# List IPAM IPs belonging to a specific Private Network
list "scaleway_ipam_ip" "by_pn" {
  provider = scaleway

  config {
    regions            = ["fr-par"]
    private_network_id = "11111111-1111-1111-1111-111111111111"
  }
}
