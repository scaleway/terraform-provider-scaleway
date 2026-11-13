resource "scaleway_iam_ssh_key" "key" {
  name       = "instance-tmpl-admin-ssh-key"
  public_key = file("~/.ssh/id_rsa.pub")
}

resource "scaleway_instance_template" "main" {
  name        = "instance-template-windows"
  server_type = "POP2-4C-16G-WIN"

  windows_rdp_ssh_key_id = scaleway_iam_ssh_key.key.id
}
