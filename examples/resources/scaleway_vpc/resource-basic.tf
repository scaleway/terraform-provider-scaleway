### Basic

resource "scaleway_vpc" "vpc01" {
  name = "my-vpc"
  tags = ["demo", "terraform"]
}
