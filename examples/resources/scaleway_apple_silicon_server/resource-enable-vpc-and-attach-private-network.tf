### Enable VPC and attach private network

resource "scaleway_vpc" "vpc-apple-silicon" {
  name = "vpc-apple-silicon"
}

resource "scaleway_vpc_private_network" "pn-apple-silicon" {
  name   = "pn-apple-silicon"
  vpc_id = scaleway_vpc.vpc-apple-silicon.id
}

resource "scaleway_apple_silicon_server" "my-server" {
  name       = "TestAccServerEnableVPC"
  type       = "M2-M"
  enable_vpc = true
  private_network {
    id = scaleway_vpc_private_network.pn-apple-silicon.id
  }
}
