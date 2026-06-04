# List Load Balancers across all zones filtered by name
list "scaleway_lb" "by_name" {
  provider = scaleway

  config {
    zones = ["*"]
    name  = "my-lb"
  }
}
