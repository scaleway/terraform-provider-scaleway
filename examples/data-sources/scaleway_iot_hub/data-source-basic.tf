### Basic

# Get info by name
data "scaleway_iot_hub" "my_hub" {
  name = "foobar"
}

# Get info by hub ID
data "scaleway_iot_hub" "my_hub" {
  hub_id = "11111111-1111-1111-1111-111111111111"
}
