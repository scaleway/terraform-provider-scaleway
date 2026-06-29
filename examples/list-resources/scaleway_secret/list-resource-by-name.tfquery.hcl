// List secrets filtered by name
list "scaleway_secret" "by_name" {
  provider = scaleway

  config {
    name = "my-secret"
  }
}
