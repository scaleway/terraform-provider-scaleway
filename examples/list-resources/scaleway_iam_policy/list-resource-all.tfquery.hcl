// List all policies in an organization
list "scaleway_iam_policy" "all" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
  }
}
