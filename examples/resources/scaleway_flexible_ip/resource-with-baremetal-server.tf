### With baremetal server

resource "scaleway_account_ssh_key" "main" {
  name       = "main"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAILHy/M5FVm5ydLGcal3e5LNcfTalbeN7QL/ZGCvDEdqJ foobar@example.com"
}

data "scaleway_baremetal_os" "by_id" {
  zone    = "fr-par-2"
  name    = "Ubuntu"
  version = "20.04 LTS (Focal Fossa)"
}

data "scaleway_baremetal_offer" "my_offer" {
  zone = "fr-par-2"
  name = "EM-A210R-HDD"
}

resource "scaleway_baremetal_server" "base" {
  zone        = "fr-par-2"
  offer       = data.scaleway_baremetal_offer.my_offer.offer_id
  os          = data.scaleway_baremetal_os.by_id.os_id
  ssh_key_ids = scaleway_account_ssh_key.main.id
}

resource "scaleway_flexible_ip" "main" {
  server_id = scaleway_baremetal_server.base.id
  zone      = "fr-par-2"
}
