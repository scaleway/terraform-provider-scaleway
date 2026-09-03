## Basic

# Get info by IP address
data "scaleway_flexible_ip" "with_ip" {
  ip_address = "1.2.3.4"
}

# Get info by IP ID
data "scaleway_flexible_ip" "with_id" {
  flexible_ip_id = "11111111-1111-1111-1111-111111111111"
}
