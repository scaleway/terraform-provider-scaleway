# List VPCs in a specific region (fr-par) for a specific project
list "scaleway_vpc" "region" {
  provider = scaleway

  config {
    regions     = ["fr-par"]
    project_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
