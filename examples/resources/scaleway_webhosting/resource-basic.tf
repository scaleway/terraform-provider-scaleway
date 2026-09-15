### Basic

data "scaleway_webhosting_offer" "by_name" {
  name          = "lite"
  control_panel = "Cpanel"
}

resource "scaleway_webhosting" "main" {
  offer_id = data.scaleway_webhosting_offer.by_name.offer_id
  email    = "your@email.com"
  domain   = "yourdomain.com"
  tags     = ["webhosting", "provider", "terraform"]
}
