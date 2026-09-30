### Basic

resource "scaleway_iot_hub" "main" {
  name         = "test-iot"
  product_plan = "plan_shared"
}

resource "scaleway_iot_device" "main" {
  hub_id = scaleway_iot_hub.main.id
  name   = "test-iot"
}
