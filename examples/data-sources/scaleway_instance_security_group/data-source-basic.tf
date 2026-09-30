# Get info by security group name
data "scaleway_instance_security_group" "my_key" {
  name = "my-security-group-name"
}

# Get info by security group id
data "scaleway_instance_security_group" "my_key" {
  security_group_id = "11111111-1111-1111-1111-111111111111"
}
