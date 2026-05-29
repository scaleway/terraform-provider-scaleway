// List all versions of all secrets in a region
list "scaleway_secret_version" "all_secrets" {
  provider = scaleway

  config {
    secret_ids = ["*"]
  }
}
