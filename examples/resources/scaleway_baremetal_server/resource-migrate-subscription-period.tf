### Migrate from hourly to monthly plan

#### Hourly Plan Example

data "scaleway_baremetal_offer" "my_offer" {
  zone                = "fr-par-2"
  name                = "EM-B112X-SSD"
  subscription_period = "hourly"
}

resource "scaleway_baremetal_server" "my_server" {
  name                     = "UpdateSubscriptionPeriod"
  offer                    = data.scaleway_baremetal_offer.my_offer.offer_id
  zone                     = "fr-par-2"
  install_config_afterward = true
}

#### Monthly Plan Example

data "scaleway_baremetal_offer" "my_offer" {
  zone                = "fr-par-2"
  name                = "EM-B112X-SSD"
  subscription_period = "monthly"
}

resource "scaleway_baremetal_server" "my_server" {
  name                     = "UpdateSubscriptionPeriod"
  offer                    = data.scaleway_baremetal_offer.my_offer.offer_id
  zone                     = "fr-par-2"
  install_config_afterward = true
}
