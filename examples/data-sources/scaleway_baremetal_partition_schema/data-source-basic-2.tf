## Basic

data "scaleway_baremetal_offer" "my_offer" {
  zone = "fr-par-1"
  name = "EM-B220E-NVME"
}

data "scaleway_baremetal_os" "my_os" {
  zone    = "fr-par-1"
  name    = "Ubuntu"
  version = "22.04 LTS (Jammy Jellyfish)"
}

resource "scaleway_iam_ssh_key" "main" {
  name       = "my-ssh-key"
  public_key = "my-ssh-key-public"
}

data "scaleway_baremetal_easy_partitioning" "test" {
  offer_id         = data.scaleway_baremetal_offer.my_offer.offer_id
  os_id            = data.scaleway_baremetal_os.my_os.os_id
  swap             = false
  ext_4_mountpoint = "/hello"
}

resource "scaleway_baremetal_server" "base" {
  name         = "my-baremetal-server"
  zone         = "fr-par-1"
  description  = "test a description"
  offer        = data.scaleway_baremetal_offer.my_offer.offer_id
  os           = data.scaleway_baremetal_os.my_os.os_id
  partitioning = data.scaleway_baremetal_easy_partitioning.test.json_partition
  tags         = ["terraform-test", "scaleway_baremetal_server", "minimal", "edited"]
  ssh_key_ids  = [scaleway_iam_ssh_key.main.id]
}
