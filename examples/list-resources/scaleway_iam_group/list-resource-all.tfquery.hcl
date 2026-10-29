# List all groups in an organization
list "scaleway_iam_group" "all" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
  }
}
