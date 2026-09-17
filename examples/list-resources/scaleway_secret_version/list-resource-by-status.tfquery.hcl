// List enabled versions of a secret
list "scaleway_secret_version" "enabled" {
  provider = scaleway

  config {
    secret_ids = [scaleway_secret.my_secret.id]
    status     = ["enabled"]
  }
}
