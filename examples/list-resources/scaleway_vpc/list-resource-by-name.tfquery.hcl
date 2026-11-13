# List VPCs across all regions filtered by name prefix (matches VPCs with names starting with "test-vpc")
list "scaleway_vpc" "by_name" {
  provider = scaleway

  config {
    regions = ["*"]
    name    = "test-vpc"
  }
}
