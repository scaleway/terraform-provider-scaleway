## Basic

# Get info by offer domain
data "scaleway_webhosting" "by_domain" {
  domain = "foobar.com"
}

# Get info by id
data "scaleway_webhosting" "by_id" {
  webhosting_id = "11111111-1111-1111-1111-111111111111"
}
