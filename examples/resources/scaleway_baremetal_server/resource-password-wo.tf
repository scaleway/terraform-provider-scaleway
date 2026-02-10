### Creating a Baremetal server using a Write Only password (not stored in state)

## Generate an ephemeral password (not stored in the state)
ephemeral "random_password" "server_password" {
  length      = 20
  special     = true
  upper       = true
  lower       = true
  numeric     = true
  min_upper   = 1
  min_lower   = 1
  min_numeric = 1
  min_special = 1
  # Exclude characters that might cause issues in some contexts
  override_special = "!@#$%^&*()_+-=[]{}|;:,.<>?"
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

resource "scaleway_iam_ssh_key" "main" {
  name       = "my_ssh_key"
  public_key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIM7HUxRyQtB2rnlhQUcbDGCZcTJg7OvoznOiyC9W6IxH user@example.com"
}

resource "scaleway_baremetal_server" "password_wo_server" {
  name                        = "test-bm-password-wo"
  zone                        = "fr-par-2"
  offer                       = data.scaleway_baremetal_offer.my_offer.offer_id
  description                 = "Baremetal server with write-only password"
  os                          = data.scaleway_baremetal_os.my_os.id
  hostname                    = "test-bm-password-wo"
  user                        = "myuser"
  password_wo                 = ephemeral.random_password.server_password.result
  password_wo_version         = 1
  service_user                = "myserviceuser"
  service_password_wo         = ephemeral.random_password.server_password.result
  service_password_wo_version = 1
  ssh_key_ids                 = [scaleway_iam_ssh_key.main.id]
}
