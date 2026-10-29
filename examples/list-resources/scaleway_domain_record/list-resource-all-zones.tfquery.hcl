// List DNS zone records across all zones in a project
list "scaleway_domain_record" "all_zones" {
  provider = scaleway

  config {
    project_ids = ["11111111-1111-1111-1111-111111111111"]
    dns_zones   = ["*"]
  }
}
