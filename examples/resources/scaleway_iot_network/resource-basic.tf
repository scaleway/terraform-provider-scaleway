### Basic

resource "scaleway_iot_network" "main" {
  name   = "main"
  hub_id = scaleway_iot_hub.main.id
  type   = "sigfox"
}
resource "scaleway_iot_hub" "main" {
  name         = "main"
  product_plan = "plan_shared"
}
