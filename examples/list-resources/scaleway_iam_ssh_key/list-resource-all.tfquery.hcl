# List all SSH keys across all projects
list "scaleway_iam_ssh_key" "all" {
  provider = scaleway

  config {
    project_ids = ["*"]
  }
}
