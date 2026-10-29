# List RDB instances with scheduled maintenances
list "scaleway_rdb_instance" "with_maintenance" {
  provider = scaleway

  config {
    regions          = ["*"]
    has_maintenances = true
  }
}
