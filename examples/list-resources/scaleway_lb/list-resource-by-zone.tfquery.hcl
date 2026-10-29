# List Load Balancers in a specific zone
list "scaleway_lb" "by_zone" {
  provider = scaleway

  config {
    zones = ["fr-par-1"]
  }
}
