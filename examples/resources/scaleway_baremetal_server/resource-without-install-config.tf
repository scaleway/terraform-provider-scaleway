### Without install config

data "scaleway_baremetal_offer" "my_offer" {
  zone = "fr-par-2"
  name = "EM-B112X-SSD"
}

resource "scaleway_baremetal_server" "my_server" {
  zone                     = "fr-par-2"
  offer                    = data.scaleway_baremetal_offer.my_offer.offer_id
  install_config_afterward = true
}
