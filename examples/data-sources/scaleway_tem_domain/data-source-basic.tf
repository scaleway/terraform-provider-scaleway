## Basic

// Get info by domain name
data "scaleway_tem_domain" "my_domain" {
  name = "example.com"
}

// Get info by domain ID
data "scaleway_tem_domain" "my_domain" {
  domain_id = "11111111-1111-1111-1111-111111111111"
}
