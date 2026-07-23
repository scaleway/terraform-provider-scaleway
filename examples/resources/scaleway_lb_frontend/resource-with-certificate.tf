## With Certificate

resource "scaleway_lb_ip" "ip01" {}

resource "scaleway_lb" "lb01" {
  ip_id = scaleway_lb_ip.ip01.id
  name  = "test-lb"
  type  = "lb-s"
}

resource "scaleway_lb_backend" "bkd01" {
  lb_id            = scaleway_lb.lb01.id
  forward_protocol = "tcp"
  forward_port     = 443
  proxy_protocol   = "none"
}

resource "scaleway_lb_certificate" "cert01" {
  lb_id = scaleway_lb.lb01.id
  name  = "test-cert-front-end"
  letsencrypt {
    common_name = "${replace(scaleway_lb_ip.ip01.ip_address, ".", "-")}.lb.${scaleway_lb.lb01.region}.scw.cloud"
  }
  # Make sure the new certificate is created before the old one can be replaced
  lifecycle {
    create_before_destroy = true
  }
}

resource "scaleway_lb_frontend" "frt01" {
  lb_id           = scaleway_lb.lb01.id
  backend_id      = scaleway_lb_backend.bkd01.id
  inbound_port    = 443
  certificate_ids = [scaleway_lb_certificate.cert01.id]
}
