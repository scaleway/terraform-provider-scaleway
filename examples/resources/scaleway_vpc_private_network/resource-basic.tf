### Basic

resource "scaleway_vpc_private_network" "pn_priv" {
  name = "subnet_demo"
  tags = ["demo", "terraform"]
}
