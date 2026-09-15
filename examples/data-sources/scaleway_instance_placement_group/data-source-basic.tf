# Get info by placement group name
data "scaleway_instance_placement_group" "my_key" {
  name = "my-placement-group-name"
}

# Get info by placement group id
data "scaleway_instance_placement_group" "my_key" {
  placement_group_id = "11111111-1111-1111-1111-111111111111"
}
