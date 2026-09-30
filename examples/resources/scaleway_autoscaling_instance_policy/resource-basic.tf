### Basic

resource "scaleway_autoscaling_instance_policy" "up" {
  instance_group_id = scaleway_autoscaling_instance_group.main.id
  name              = "scale-up-if-cpu-high"
  action            = "scale_up"
  type              = "flat_count"
  value             = 1
  priority          = 1

  metric {
    name               = "cpu scale up"
    managed_metric     = "managed_metric_instance_cpu"
    operator           = "operator_greater_than"
    aggregate          = "aggregate_average"
    sampling_range_min = 5
    threshold          = 70
  }
}

resource "scaleway_autoscaling_instance_policy" "down" {
  instance_group_id = scaleway_autoscaling_instance_group.main.id
  name              = "scale-down-if-cpu-low"
  action            = "scale_down"
  type              = "flat_count"
  value             = 1
  priority          = 2

  metric {
    name               = "cpu scale down"
    managed_metric     = "managed_metric_instance_cpu"
    operator           = "operator_less_than"
    aggregate          = "aggregate_average"
    sampling_range_min = 5
    threshold          = 40
  }
}
