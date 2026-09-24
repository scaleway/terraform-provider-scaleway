### Let's Encrypt

resource "scaleway_lb_certificate" "cert01" {
  lb_id = scaleway_lb.lb01.id
  name  = "cert1"

  letsencrypt {
    common_name = "example.org"
    subject_alternative_name = [
      "sub1.example.com",
      "sub2.example.com"
    ]
  }
  # Make sure the new certificate is created before the old one can be replaced
  lifecycle {
    create_before_destroy = true
  }
}
