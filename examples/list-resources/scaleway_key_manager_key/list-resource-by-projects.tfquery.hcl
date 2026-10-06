# List keys in specific projects
list "scaleway_keymanager_key" "by_projects" {
  provider = scaleway

  config {
    project_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
  }
}
