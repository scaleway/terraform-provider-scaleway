### Basic

# Get all Projects in an Organization
data "scaleway_account_projects" "all" {
  organization_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
