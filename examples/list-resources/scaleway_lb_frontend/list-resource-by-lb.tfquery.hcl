# List all frontends for a specific Load Balancer
list "scaleway_lb_frontend" "by_lb" {
  provider = scaleway

  config {
    zones  = ["fr-par-1"]
    lb_ids = ["11111111-1111-1111-1111-111111111111"]
  }
}
