### Basic

resource "scaleway_iot_hub" "main" {
  name         = "test-iot"
  product_plan = "plan_shared"
}
