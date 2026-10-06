// List projects filtered by name
list "scaleway_account_project" "by_name" {
  provider = scaleway

  config {
    name = "my-project"
  }
}
