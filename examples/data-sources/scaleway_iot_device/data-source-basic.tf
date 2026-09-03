### Basic

# Get info by name 
data "scaleway_iot_device" "my_device" {
  name = "foobar"
}

# Get info by name and hub_id
data "scaleway_iot_device" "my_device" {
  name   = "foobar"
  hub_id = "11111111-1111-1111-1111-111111111111"
}

# Get info by device ID
data "scaleway_iot_device" "my_device" {
  device_id = "11111111-1111-1111-1111-111111111111"
}
