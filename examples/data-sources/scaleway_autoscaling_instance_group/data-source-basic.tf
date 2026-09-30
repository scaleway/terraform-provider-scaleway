## Basic

# Get info by name
data "scaleway_autoscaling_instance_group" "by_name" {
  name = "my-instance-group"
}

# Get info by ID
data "scaleway_autoscaling_instance_group" "by_id" {
  instance_group_id = "11111111-1111-1111-1111-111111111111"
}
