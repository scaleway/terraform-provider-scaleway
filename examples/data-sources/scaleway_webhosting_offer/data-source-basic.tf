## Basic

# Get info by offer name
data "scaleway_webhosting_offer" "by_name" {
  name          = "performance"
  control_panel = "Cpanel"
}

# Get info by offer id
data "scaleway_webhosting_offer" "by_id" {
  offer_id = "de2426b4-a9e9-11ec-b909-0242ac120002"
}
