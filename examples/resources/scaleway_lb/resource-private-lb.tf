### Private LB

resource "scaleway_lb" "base" {
  name               = "private-lb"
  type               = "LB-S"
  assign_flexible_ip = false
}
