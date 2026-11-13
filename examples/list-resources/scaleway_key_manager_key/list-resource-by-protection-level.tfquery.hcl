# List keys filtered by protection level
list "scaleway_key_manager_key" "by_protection_level" {
  provider = scaleway

  config {
    protection_level = "hsm"
  }
}
