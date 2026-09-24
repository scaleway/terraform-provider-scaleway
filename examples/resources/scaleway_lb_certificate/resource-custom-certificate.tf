### Custom Certificate

resource "scaleway_lb_certificate" "cert01" {
  lb_id = scaleway_lb.lb01.id
  name  = "custom-cert"
  custom_certificate {
    certificate_chain = <<EOF
CERTIFICATE_CHAIN_CONTENTS
EOF
  }
}
