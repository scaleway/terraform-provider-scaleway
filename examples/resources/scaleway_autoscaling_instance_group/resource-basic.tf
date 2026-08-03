### Basic

resource "scaleway_autoscaling_instance_group" "main" {
  name        = "asg-group"
  template_id = scaleway_autoscaling_instance_template.main.id
  tags        = ["terraform-test", "instance-group"]
  capacity {
    max_replicas   = 5
    min_replicas   = 1
    cooldown_delay = "300"
  }
  load_balancer {
    id                 = scaleway_lb.main.id
    backend_ids        = [scaleway_lb_backend.main.id]
    private_network_id = scaleway_vpc_private_network.main.id
  }
}
