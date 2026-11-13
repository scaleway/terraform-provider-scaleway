# List disabled SSH keys in specific projects
list "scaleway_iam_ssh_key" "disabled" {
  provider = scaleway

  config {
    project_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
    disabled = true
  }
}
