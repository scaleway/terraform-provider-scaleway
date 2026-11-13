# List SSH keys filtered by name
list "scaleway_iam_ssh_key" "by_name" {
  provider = scaleway

  config {
    name = "my-ssh-key"
  }
}
