### Basic

# Get info by IP address
data "scaleway_lb_ip" "my_ip" {
  ip_address = "0.0.0.0"
}

# Get info by IP ID
data "scaleway_lb_ip" "my_ip" {
  ip_id = "11111111-1111-1111-1111-111111111111"
}
