### Create a DNS zone

resource "scaleway_domain_zone" "test" {
  domain    = "scaleway-terraform.com"
  subdomain = "test"
}
