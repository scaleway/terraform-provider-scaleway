### Example Basic

resource "scaleway_vpc" "vpc01" {
  name = "my-vpc-source"
}

resource "scaleway_vpc" "vpc02" {
  name = "my-vpc-target"
}

resource "scaleway_vpc_connector" "main" {
  name          = "my-vpc-connector"
  vpc_id        = scaleway_vpc.vpc01.id
  target_vpc_id = scaleway_vpc.vpc02.id
}
