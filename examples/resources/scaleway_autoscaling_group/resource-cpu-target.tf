resource "scaleway_instance_template" "main" {
  server_type = "GP1-L"
}

resource "scaleway_autoscaling_group" "main" {
  template_id = scaleway_instance_template.main.id
  tags        = ["tf-examples", "scaleway_autoscaling_group", "cpu-target"]
  name        = "asg-cpu-target"

  scaling_policy = {
    minimum_size       = 0
    maximum_size       = 25
    scale_in_cooldown  = "2m"
    scale_out_cooldown = "1m"
    scale_in_step      = 5
    scale_out_step     = 5

    cpu_target = 92
  }
}
