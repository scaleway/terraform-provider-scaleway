// List DNS zone records in a specific zone
list "scaleway_domain_record" "by_dns_zone" {
  provider = scaleway

  config {
    dns_zones = ["www.example.com"]
    type      = "A"
  }
}
