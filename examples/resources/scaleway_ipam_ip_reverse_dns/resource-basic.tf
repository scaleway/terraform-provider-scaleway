### Basic

resource "scaleway_instance_ip" "ip01" {
  type = "routed_ipv6"
}

resource "scaleway_instance_server" "srv01" {
  name   = "tf-tests-instance-server-ips"
  ip_ids = [scaleway_instance_ip.ip01.id]
  image  = "ubuntu_jammy"
  type   = "PRO2-XXS"
  state  = "stopped"
}

data "scaleway_ipam_ip" "ipam01" {
  resource {
    id   = scaleway_instance_server.srv01.id
    type = "instance_server"
  }
  type = "ipv6"
}

resource "scaleway_domain_record" "tf_AAAA" {
  dns_zone = "example.com"
  name     = ""
  type     = "AAAA"
  data     = cidrhost(data.scaleway_ipam_ip.ipam01.address_cidr, 42)
  ttl      = 3600
  priority = 1
}

resource "scaleway_ipam_ip_reverse_dns" "base" {
  ipam_ip_id = data.scaleway_ipam_ip.ipam01.id

  hostname = "example.com"
  address  = cidrhost(data.scaleway_ipam_ip.ipam01.address_cidr, 42)
}
