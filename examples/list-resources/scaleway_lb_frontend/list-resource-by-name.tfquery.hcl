# List frontends filtered by name
list "scaleway_lb_frontend" "by_name" {
  provider = scaleway

  config {
    zones  = ["fr-par-1"]
    lb_ids = ["11111111-1111-1111-1111-111111111111"]
    name   = "my-frontend"
  }
}
