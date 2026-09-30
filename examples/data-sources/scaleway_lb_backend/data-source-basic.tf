### Basic

resource "scaleway_lb_ip" "main" {
}

resource "scaleway_lb" "main" {
  ip_id = scaleway_lb_ip.main.id
  name  = "data-test-lb-backend"
  type  = "LB-S"
}

resource "scaleway_lb_backend" "main" {
  lb_id            = scaleway_lb.main.id
  name             = "backend01"
  forward_protocol = "http"
  forward_port     = "80"
}

data "scaleway_lb_backend" "byID" {
  backend_id = scaleway_lb_backend.main.id
}

data "scaleway_lb_backend" "byName" {
  name  = scaleway_lb_backend.main.name
  lb_id = scaleway_lb.main.id
}
