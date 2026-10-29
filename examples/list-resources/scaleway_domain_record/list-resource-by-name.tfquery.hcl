// List DNS zone records filtered by name
list "scaleway_domain_record" "by_name" {
  provider = scaleway

  config {
    dns_zones = ["www.example.com"]
    name      = "www"
    type      = "A"
  }
}
