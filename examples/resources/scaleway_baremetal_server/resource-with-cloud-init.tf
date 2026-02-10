### With cloud-init

data "scaleway_iam_ssh_key" "my_ssh_key" {
  name = "main"
}

data "scaleway_baremetal_os" "my_os" {
  zone    = "fr-par-1"
  name    = "Ubuntu"
  version = "22.04 LTS (Jammy Jellyfish)"
}


data "scaleway_baremetal_offer" "my_offer" {
  zone = "fr-par-2"
  name = "EM-I220E-NVME"
}

resource "scaleway_baremetal_server" "my_server_ci" {
  zone        = "fr-par-2"
  offer       = data.scaleway_baremetal_offer.my_offer.offer_id
  os          = data.scaleway_baremetal_os.my_os.os_id
  cloud_init  = "#cloud-config\napt_update: true\napt_upgrade: true"
  ssh_key_ids = [data.scaleway_iam_ssh_key.my_ssh_key.id]
}
