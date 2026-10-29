# VPC resources
resource "scaleway_vpc" "vpc" {}
resource "scaleway_vpc_private_network" "pn" {
  vpc_id = scaleway_vpc.vpc.id
}

# LB resources
resource "scaleway_lb" "lb" {
  type = "LB-S"
}
resource "scaleway_lb_backend" "back0" {
  lb_id            = scaleway_lb.lb.id
  forward_protocol = "http"
  forward_port     = 80
}
resource "scaleway_lb_backend" "back1" {
  lb_id            = scaleway_lb.lb.id
  forward_protocol = "tcp"
  forward_port     = 81
}

# Template
resource "scaleway_instance_template" "main" {
  tags        = ["terraform-test", "scaleway_autoscaling_group", "lb_config"]
  server_type = "PRO2-M"
}

# AutoScaling Group
resource "scaleway_autoscaling_group" "main" {
  template_id = scaleway_instance_template.main.id

  scaling_policy = {
    minimum_size  = 0
    maximum_size  = 25
    memory_target = 75
  }

  load_balancer_configuration = {
    load_balancer_id = scaleway_lb.lb.id
    backends = [{
      backend_id     = scaleway_lb_backend.back0.id
      address_family = "ipv6"
      }, {
      backend_id         = scaleway_lb_backend.back1.id
      address_family     = "ipv4"
      private_network_id = scaleway_vpc_private_network.pn.id
    }]
    auto_healing = {
      enabled      = true
      grace_period = "10m"
    }
  }
}
