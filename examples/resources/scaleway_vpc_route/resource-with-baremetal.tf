### With Baremetal

resource "scaleway_vpc" "vpc01" {
  name = "tf-vpc-vpn"
}

resource "scaleway_vpc_private_network" "pn01" {
  name = "tf-pn-vpn"
  ipv4_subnet {
    subnet = "172.16.64.0/22"
  }
  vpc_id = scaleway_vpc.vpc01.id
}

data "scaleway_baremetal_os" "my_os" {
  zone    = "fr-par-2"
  name    = "Ubuntu"
  version = "22.04 LTS (Jammy Jellyfish)"
}

data "scaleway_baremetal_offer" "my_offer" {
  zone = "fr-par-2"
  name = "EM-B112X-SSD"
}

data "scaleway_baremetal_option" "private_network" {
  zone = "fr-par-2"
  name = "Private Network"
}

data "scaleway_iam_ssh_key" "my_key" {
  name = "main"
}

resource "scaleway_baremetal_server" "my_server" {
  zone        = "fr-par-2"
  offer       = data.scaleway_baremetal_offer.my_offer.offer_id
  os          = data.scaleway_baremetal_os.my_os.os_id
  ssh_key_ids = [data.scaleway_iam_ssh_key.my_key.id]

  options {
    id = data.scaleway_baremetal_option.private_network.option_id
  }
  private_network {
    id = scaleway_vpc_private_network.pn01.id
  }
}

resource "scaleway_vpc_route" "rt01" {
  vpc_id              = scaleway_vpc.vpc01.id
  description         = "tf-route-vpn"
  tags                = ["tf", "route"]
  destination         = "10.0.0.0/24"
  nexthop_resource_id = scaleway_baremetal_server.my_server.private_network.0.mapping_id
}
