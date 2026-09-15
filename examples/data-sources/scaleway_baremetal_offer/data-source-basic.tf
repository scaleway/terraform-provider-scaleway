## Basic

# Get info by offer name
data "scaleway_baremetal_offer" "my_offer" {
  zone = "fr-par-2"
  name = "EM-A210R-SATA"
}

# Get info by offer id
data "scaleway_baremetal_offer" "my_offer" {
  zone     = "fr-par-2"
  offer_id = "25dcf38b-c90c-4b18-97a2-6956e9d1e113"
}
