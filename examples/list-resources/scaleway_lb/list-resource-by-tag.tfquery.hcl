# List Load Balancers filtered by tag
list "scaleway_lb" "by_tag" {
  provider = scaleway

  config {
    zones = ["*"]
    tags  = ["production"]
  }
}
