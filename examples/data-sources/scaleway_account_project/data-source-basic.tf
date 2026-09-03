### Basic

# Get info by name
data "scaleway_account_project" "by_name" {
  name            = "myproject"
  organization_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
# Get default project
data "scaleway_account_project" "by_name" {
  name = "default"
}
# Get info by ID
data "scaleway_account_project" "by_id" {
  project_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
