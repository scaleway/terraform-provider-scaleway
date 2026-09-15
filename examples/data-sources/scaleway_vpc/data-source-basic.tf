### Basic

# Get info by name
data "scaleway_vpc" "by_name" {
  name = "foobar"
}

# Get info by ID
data "scaleway_vpc" "by_id" {
  vpc_id = "11111111-1111-1111-1111-111111111111"
}

# Get default VPC info
data "scaleway_vpc" "default" {
  is_default = true
}
