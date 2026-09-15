### Basic

# Get info by name
data "scaleway_vpc_private_network" "my_name" {
  name = "foobar"
}

# Get info by name and VPC ID
data "scaleway_vpc_private_network" "my_name_and_vpc_id" {
  name   = "foobar"
  vpc_id = "11111111-1111-1111-1111-111111111111"
}

# Get info by name in a specific region
data "scaleway_vpc_private_network" "my_name_and_region" {
  name   = "foobar"
  region = "nl-ams"
}

# Get info by Private Network ID
data "scaleway_vpc_private_network" "my_id" {
  private_network_id = "11111111-1111-1111-1111-111111111111"
}
