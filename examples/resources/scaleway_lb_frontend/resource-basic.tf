### Basic

resource "scaleway_lb_frontend" "frontend01" {
  lb_id        = scaleway_lb.lb01.id
  backend_id   = scaleway_lb_backend.backend01.id
  name         = "frontend01"
  inbound_port = "80"
}
