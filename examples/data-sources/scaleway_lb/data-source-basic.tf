### Basic

# Get info by name
data "scaleway_lb" "by_name" {
  name = "foobar"
}

# Get info by ID
data "scaleway_lb" "by_id" {
  lb_id = "11111111-1111-1111-1111-111111111111"
}
