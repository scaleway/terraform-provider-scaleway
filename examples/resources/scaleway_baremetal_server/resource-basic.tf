### Basic

data "scaleway_iam_ssh_key" "my_ssh_key" {
  name       = "main"
  public_key = "ssh XXXXXXXXXXX"
}

data "scaleway_baremetal_offer" "my_offer" {
  zone = "fr-par-2"
  name = "EM-I220E-NVME"
}

resource "scaleway_baremetal_server" "my_server" {
  zone        = "fr-par-2"
  offer       = data.scaleway_baremetal_offer.my_offer.offer_id
  os          = "d17d6872-0412-45d9-a198-af82c34d3c5c"
  ssh_key_ids = [data.scaleway_iam_ssh_key.my_ssh_key.id]
}
