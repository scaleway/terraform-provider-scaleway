### Basic

resource "scaleway_vpc_public_gateway" "main" {
  name = "public_gateway_demo"
  type = "VPC-GW-S"
  tags = ["demo", "terraform"]
}
