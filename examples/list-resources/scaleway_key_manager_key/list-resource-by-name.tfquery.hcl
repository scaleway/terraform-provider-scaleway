# List keys filtered by name
list "scaleway_keymanager_key" "by_name" {
  provider = scaleway

  config {
    name = "my-key"
  }
}
