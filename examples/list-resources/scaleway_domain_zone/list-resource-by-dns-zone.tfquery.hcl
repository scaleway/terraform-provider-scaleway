// List a specific DNS zone by FQDN
list "scaleway_domain_zone" "by_dns_zone" {
  provider = scaleway

  config {
    domains   = ["example.com"]
    dns_zones = ["www.example.com"]
  }
}
