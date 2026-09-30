### Basic

resource "scaleway_vpc_public_gateway" "main" {
  name = "demo"
  type = "VPC-GW-S"
  zone = "nl-ams-1"
}

data "scaleway_vpc_public_gateway" "pg_test_by_name" {
  name = scaleway_vpc_public_gateway.main.name
  zone = "nl-ams-1"
}

data "scaleway_vpc_public_gateway" "pg_test_by_id" {
  public_gateway_id = scaleway_vpc_public_gateway.main.id
}
