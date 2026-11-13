### With private network

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

resource "scaleway_vpc_private_network" "pn" {
  name = "baremetal_private_network"
}

resource "scaleway_iam_ssh_key" "my_ssh_key" {
  name       = "main"
  public_key = "ssh XXXXXXXXXXX"
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
    id = scaleway_vpc_private_network.pn.id
  }
}
