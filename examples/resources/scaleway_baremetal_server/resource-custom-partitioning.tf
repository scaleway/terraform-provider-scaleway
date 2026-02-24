### With custom partitioning

data "scaleway_iam_ssh_key" "my_ssh_key" {
  name       = "main"
  public_key = "ssh XXXXXXXXXXX"
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

resource "scaleway_baremetal_server" "my_server" {
  zone         = "fr-par-2"
  offer        = data.scaleway_baremetal_offer.my_offer.offer_id
  os           = data.scaleway_baremetal_os.my_os.os_id
  partitioning = "{\"disks\":[{\"device\":\"/dev/nvme0n1\",\"partitions\":[{\"label\":\"uefi\",\"number\":1,\"size\":536870912},{\"label\":\"swap\",\"number\":2,\"size\":4294967296},{\"label\":\"boot\",\"number\":3,\"size\":1073741824},{\"label\":\"root\",\"number\":4,\"size\":1017827045376}]},{\"device\":\"/dev/nvme1n1\",\"partitions\":[{\"label\":\"swap\",\"number\":1,\"size\":4294967296},{\"label\":\"boot\",\"number\":2,\"size\":1073741824},{\"label\":\"root\",\"number\":3,\"size\":1017827045376}]}],\"filesystems\":[{\"device\":\"/dev/nvme0n1p1\",\"format\":\"fat32\",\"mountpoint\":\"/boot/efi\"},{\"device\":\"/dev/md0\",\"format\":\"ext4\",\"mountpoint\":\"/boot\"},{\"device\":\"/dev/md1\",\"format\":\"ext4\",\"mountpoint\":\"/\"}],\"raids\":[{\"devices\":[\"/dev/nvme0n1p3\",\"/dev/nvme1n1p2\"],\"level\":\"raid_level_1\",\"name\":\"/dev/md0\"},{\"devices\":[\"/dev/nvme0n1p4\",\"/dev/nvme1n1p3\"],\"level\":\"raid_level_1\",\"name\":\"/dev/md1\"}],\"zfs\":{\"pools\":[]}}"
  ssh_key_ids  = [data.scaleway_iam_ssh_key.my_ssh_key.id]
}
