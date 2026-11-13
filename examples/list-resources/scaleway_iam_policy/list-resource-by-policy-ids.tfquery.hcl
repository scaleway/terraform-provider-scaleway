// List policies by policy IDs
list "scaleway_iam_policy" "by_ids" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
    policy_ids      = ["11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"]
  }
}
