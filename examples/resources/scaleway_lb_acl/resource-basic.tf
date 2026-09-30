### Basic

resource "scaleway_lb_acl" "acl01" {
  frontend_id = scaleway_lb_frontend.frt01.id
  name        = "acl01"
  description = "Exclude well-known IPs"
  index       = 0
  # Allow downstream requests from: 192.168.0.1, 192.168.0.2 or 192.168.10.0/24
  action {
    type = "allow"
  }
  match {
    ip_subnet = ["192.168.0.1", "192.168.0.2", "192.168.10.0/24"]
  }
}
