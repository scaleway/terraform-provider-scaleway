// List all projects in an organization
list "scaleway_account_project" "all" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
  }
}
