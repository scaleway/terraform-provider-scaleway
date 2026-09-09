// List all versions of a secret
list "scaleway_secret_version" "all" {
  provider = scaleway

  config {
    secret_ids = [scaleway_secret.my_secret.id]
  }
}
