# List VPCs in all regions for the default project filtered by a specific tag
list "scaleway_vpc" "by_tag" {
  provider = scaleway

  config {
    regions     = ["*"]
    tags        = ["foobar"]
  }
}
