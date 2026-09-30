### With custom certificate

resource "scaleway_iot_hub" "main" {
  name         = "test-iot"
  product_plan = "plan_shared"
}

data "local_file" "device_cert" {
  filename = "device-certificate.pem"
}

resource "scaleway_iot_device" "main" {
  hub_id = scaleway_iot_hub.main.id
  name   = "test-iot"
  certificate {
    crt = data.local_file.device_cert.content
  }
}
