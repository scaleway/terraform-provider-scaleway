### IPAM IP ID

# Get info by ipam ip id
data "scaleway_ipam_ip" "by_id" {
  ipam_ip_id = "11111111-1111-1111-1111-111111111111"
}
