### With IPAM IP IDs

resource "scaleway_vpc" "vpc01" {
  name = "TestAccScalewayBaremetalIPAM"
}

resource "scaleway_vpc_private_network" "pn01" {
  name = "TestAccScalewayBaremetalIPAM"
  ipv4_subnet {
    subnet = "172.16.64.0/22"
  }
  vpc_id = scaleway_vpc.vpc01.id
}

resource "scaleway_ipam_ip" "ip01" {
  address = "172.16.64.7"
  source {
    private_network_id = scaleway_vpc_private_network.pn01.id
  }
}

data "scaleway_iam_ssh_key" "my_ssh_key" {
  name       = "main"
  public_key = "ssh XXXXXXXXXXX"
}

data "scaleway_baremetal_os" "my_os" {
  zone    = "fr-par-1"
  name    = "Ubuntu"
  version = "22.04 LTS (Jammy Jellyfish)"
}

data "scaleway_baremetal_offer" "my_offer" {
  zone = "fr-par-1"
  name = "EM-A115X-SSD"
}

data "scaleway_baremetal_option" "private_network" {
  zone = "fr-par-1"
  name = "Private Network"
}

resource "scaleway_baremetal_server" "my_server" {
  zone        = "fr-par-2"
  offer       = data.scaleway_baremetal_offer.my_offer.offer_id
  os          = data.scaleway_baremetal_os.my_os.os_id
  ssh_key_ids = [data.scaleway_iam_ssh_key.my_ssh_key.id]

  options {
    id = data.scaleway_baremetal_option.private_network.option_id
  }
  private_network {
    id          = scaleway_vpc_private_network.pn01.id
    ipam_ip_ids = [scaleway_ipam_ip.ip01.id]
  }
}
