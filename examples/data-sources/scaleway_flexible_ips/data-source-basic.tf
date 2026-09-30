## Basic

# Find ips that share the same tags
data "scaleway_flexible_ips" "fips_by_tags" {
  tags = ["a tag"]
}

# Find ips that share the same Server ID
data "scaleway_baremetal_offer" "my_offer" {
  name = "EM-B112X-SSD"
}

resource "scaleway_baremetal_server" "base" {
  name                     = "MyServer"
  offer                    = data.scaleway_baremetal_offer.my_offer.offer_id
  install_config_afterward = true
}

resource "scaleway_flexible_ip" "first" {
  server_id = scaleway_baremetal_server.base.id
  tags      = ["foo", "first"]
}

resource "scaleway_flexible_ip" "second" {
  server_id = scaleway_baremetal_server.base.id
  tags      = ["foo", "second"]
}

data "scaleway_flexible_ips" "fips_by_server_id" {
  server_ids = [scaleway_baremetal_server.base.id]
  depends_on = [scaleway_flexible_ip.first, scaleway_flexible_ip.second]
}
