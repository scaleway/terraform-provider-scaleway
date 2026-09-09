resource "scaleway_instance_template" "main" {
  server_type = "PRO2-M"
}

resource "scaleway_autoscaling_group" "main" {
  template_id = scaleway_instance_template.main.id
  tags        = ["tf-examples", "scaleway_autoscaling_group", "fixed-size"]
  name        = "asg-fixed-size"

  scaling_policy = {
    minimum_size       = 1
    maximum_size       = 20
    scale_in_cooldown  = "7m"
    scale_out_cooldown = "30s"

    fixed_size = 10
  }
}
