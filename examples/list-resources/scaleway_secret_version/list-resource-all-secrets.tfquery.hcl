// List all versions of all secrets in all regions
list "scaleway_secret_version" "all_secrets" {
  provider = scaleway

  config {
    regions    = ["*"]
    secret_ids = ["*"]
  }
}
