### With bastion

resource "scaleway_iam_ssh_key" "key1" {
  name       = "key1"
  public_key = file("~/.ssh/id_rsa.pub")
}

resource "scaleway_iam_ssh_key" "key2" {
  name       = "key2"
  public_key = file("~/.ssh/another_key.pub")
}

# Use a local variable to compute a hash of the SSH keys
locals {
  ssh_keys_hash = sha256(join(",", [
    scaleway_iam_ssh_key.key1.public_key,
    scaleway_iam_ssh_key.key2.public_key,
  ]))
}

resource "scaleway_vpc_public_gateway" "main" {
  name             = "public_gateway_demo"
  type             = "VPC-GW-S"
  tags             = ["demo", "terraform"]
  bastion_enabled  = true
  bastion_port     = 61000
  refresh_ssh_keys = local.ssh_keys_hash
}
