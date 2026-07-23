### With IPv6

resource "scaleway_lb_ip" "v4" {
}
resource "scaleway_lb_ip" "v6" {
  is_ipv6 = true
}
resource "scaleway_lb" "main" {
  ip_ids = [scaleway_lb_ip.v4.id, scaleway_lb_ip.v6.id]
  name   = "ipv6-lb"
  type   = "LB-S"
}
