// List DNS zones across all domains in a project
list "scaleway_domain_zone" "all_domains" {
  provider = scaleway

  config {
    project_ids = ["11111111-1111-1111-1111-111111111111"]
    domains     = ["*"]
  }
}
