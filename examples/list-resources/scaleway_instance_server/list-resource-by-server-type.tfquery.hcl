# List servers filtered by server-type on the default project across all zones
list "scaleway_instance_server" "by_server_type" {
  provider = scaleway

  config {
    zones       = ["*"]
    server_type = "PRO2-L"
  }
}