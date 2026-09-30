### Rest Route

resource "scaleway_iot_route" "main" {
  name   = "main"
  hub_id = scaleway_iot_hub.main.id
  topic  = "#"
  rest {
    verb = "get"
    uri  = "http://scaleway.com"
    headers = {
      X-awesome-header = "my-awesome-value"
    }
  }
}

resource "scaleway_iot_hub" "main" {
  name         = "main"
  product_plan = "plan_shared"
}
