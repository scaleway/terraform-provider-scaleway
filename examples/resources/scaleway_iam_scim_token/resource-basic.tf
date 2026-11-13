### Basic SCIM Token creation

# Enable SCIM for your organization
resource "scaleway_iam_scim" "main" {
  organization_id = "your-organization-id"
}

# Create a SCIM token
resource "scaleway_iam_scim_token" "main" {
  scim_id         = scaleway_iam_scim.main.id
  organization_id = "your-organization-id"
}
