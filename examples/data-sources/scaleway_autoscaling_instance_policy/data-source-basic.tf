## Basic

# Get info by name (instance_group_id is required when using name)
data "scaleway_autoscaling_instance_policy" "by_name" {
  name              = "my-instance-policy"
  instance_group_id = scaleway_autoscaling_instance_group.main.id
}

# Get info by ID
data "scaleway_autoscaling_instance_policy" "by_id" {
  instance_policy_id = "11111111-1111-1111-1111-111111111111"
}
