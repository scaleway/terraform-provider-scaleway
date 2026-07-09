### With path-begin matching for HTTP backends

resource "scaleway_lb_ip" "ip" {}

resource "scaleway_lb" "lb" {
  ip_id = scaleway_lb_ip.ip.id
  name  = "my-lb"
  type  = "lb-s"
}

resource "scaleway_lb_backend" "app" {
  lb_id            = scaleway_lb.lb.id
  forward_protocol = "http"
  forward_port     = 80
  proxy_protocol   = "none"
}

resource "scaleway_lb_backend" "admin" {
  lb_id            = scaleway_lb.lb.id
  forward_protocol = "http"
  forward_port     = 8080
  proxy_protocol   = "none"
}

resource "scaleway_lb_frontend" "frontend" {
  lb_id        = scaleway_lb.lb.id
  backend_id   = scaleway_lb_backend.app.id
  inbound_port = 80
}

resource "scaleway_lb_route" "admin_route" {
  frontend_id      = scaleway_lb_frontend.frontend.id
  backend_id       = scaleway_lb_backend.admin.id
  match_path_begin = "/admin"
}

resource "scaleway_lb_route" "default_route" {
  frontend_id      = scaleway_lb_frontend.frontend.id
  backend_id       = scaleway_lb_backend.app.id
  match_path_begin = "/"
}
