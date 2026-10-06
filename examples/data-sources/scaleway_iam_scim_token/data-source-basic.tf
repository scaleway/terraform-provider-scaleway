### Get information about a SCIM token

# First, enable SCIM for your organization
resource "scaleway_iam_scim" "main" {
  organization_id = "11111111-1111-1111-1111-111111111111"
}

# Create a SCIM token (or use an existing one)
resource "scaleway_iam_scim_token" "main" {
  scim_id         = scaleway_iam_scim.main.id
  organization_id = "11111111-1111-1111-1111-111111111111"
}

# Get information about the SCIM token
data "scaleway_iam_scim_token" "main" {
  scim_id         = scaleway_iam_scim.main.id
  token_id        = scaleway_iam_scim_token.main.id
  organization_id = "11111111-1111-1111-1111-111111111111"
}
