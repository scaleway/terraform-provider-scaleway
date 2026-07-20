# List servers filtered by server-type on the default project across all zones
list "scaleway_instance_server" "by_name_and_tag" {
  provider = scaleway

  config {
    zones       = ["*"]
    server_type = "PRO2-L"
  }
}