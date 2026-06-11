### Enable IAM SCIM for an organization

resource "scaleway_iam_scim" "main" {
  organization_id = "11111111-1111-1111-1111-111111111111"
}
