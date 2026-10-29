// List DNS zones for a domain in the default project
list "scaleway_domain_zone" "by_domain" {
  provider = scaleway

  config {
    domains = ["example.com"]
  }
}
