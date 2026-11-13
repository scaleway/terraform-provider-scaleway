# List VPC routes across all regions
list "scaleway_vpc_route" "all" {
  provider = scaleway

  config {
    regions = ["*"]
  }
}
