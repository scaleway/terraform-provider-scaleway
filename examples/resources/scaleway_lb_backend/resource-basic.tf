### Basic

resource "scaleway_lb_backend" "backend01" {
  lb_id            = scaleway_lb.lb01.id
  name             = "backend01"
  forward_protocol = "http"
  forward_port     = "80"
}
