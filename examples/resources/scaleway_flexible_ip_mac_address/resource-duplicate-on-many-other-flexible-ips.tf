### Duplicate on many other flexible IPs

data "scaleway_baremetal_offer" "my_offer" {
  name = "EM-B112X-SSD"
}

resource "scaleway_baremetal_server" "base" {
  name                     = "TestAccScalewayBaremetalServer_WithoutInstallConfig"
  offer                    = data.scaleway_baremetal_offer.my_offer.offer_id
  install_config_afterward = true
}

resource "scaleway_flexible_ip" "ip01" {
  server_id = scaleway_baremetal_server.base.id
}

resource "scaleway_flexible_ip" "ip02" {
  server_id = scaleway_baremetal_server.base.id
}

resource "scaleway_flexible_ip" "ip03" {
  server_id = scaleway_baremetal_server.base.id
}

resource "scaleway_flexible_ip_mac_address" "main" {
  flexible_ip_id = scaleway_flexible_ip.ip01.id
  type           = "kvm"
  flexible_ip_ids_to_duplicate = [
    scaleway_flexible_ip.ip02.id,
    scaleway_flexible_ip.ip03.id
  ]
}
