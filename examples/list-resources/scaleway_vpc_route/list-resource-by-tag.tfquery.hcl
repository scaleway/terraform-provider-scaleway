# List VPC routes filtered by tag
list "scaleway_vpc_route" "by_tag" {
  provider = scaleway

  config {
    regions = ["*"]
    tags    = ["production"]
  }
}
