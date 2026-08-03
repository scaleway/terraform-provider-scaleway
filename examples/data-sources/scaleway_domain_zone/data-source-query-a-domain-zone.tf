### Query a domain zone

# Get zone
data "scaleway_domain_zone" "main" {
  domain    = "scaleway-terraform.com"
  subdomain = "test"
}
