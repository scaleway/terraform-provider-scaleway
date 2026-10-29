# List backends filtered by name across multiple Load Balancers
list "scaleway_lb_backend" "by_name" {
  provider = scaleway

  config {
    zones = ["fr-par-1"]
    lb_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
    name = "my-backend"
  }
}
