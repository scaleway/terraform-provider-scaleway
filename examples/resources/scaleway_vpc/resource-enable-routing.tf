### Enable routing

resource "scaleway_vpc" "vpc01" {
  name           = "my-vpc"
  tags           = ["demo", "terraform", "routing"]
  enable_routing = true
}
