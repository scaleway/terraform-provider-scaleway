// List projects filtered by project IDs
list "scaleway_account_project" "by_project_ids" {
  provider = scaleway

  config {
    project_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
  }
}
