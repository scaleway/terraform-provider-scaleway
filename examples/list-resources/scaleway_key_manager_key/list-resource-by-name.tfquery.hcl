# List keys filtered by name
list "scaleway_key_manager_key" "by_name" {
  provider = scaleway

  config {
    name = "my-key"
  }
}
