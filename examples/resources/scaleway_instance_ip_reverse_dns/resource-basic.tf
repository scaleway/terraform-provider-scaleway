resource "scaleway_instance_ip" "server_ip" {}

resource "scaleway_domain_record" "tf_A" {
  dns_zone = "scaleway.com"
  name     = "www"
  type     = "A"
  data     = scaleway_instance_ip.server_ip.address
  ttl      = 3600
  priority = 1
}

resource "scaleway_instance_ip_reverse_dns" "reverse" {
  ip_id   = scaleway_instance_ip.server_ip.id
  reverse = "www.scaleway.com"
}
