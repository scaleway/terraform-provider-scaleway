### Basic

resource "scaleway_domain_record" "tf_A" {
  dns_zone = "example.com"
  name     = "tf"
  type     = "A"
  data     = scaleway_vpc_public_gateway_ip.main.address
  ttl      = 3600
  priority = 1
}

resource "scaleway_vpc_public_gateway_ip" "main" {
  reverse = "tf.example.com"
}
