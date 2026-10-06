// List policies by user IDs
list "scaleway_iam_policy" "by_user_ids" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
    user_ids        = ["11111111-1111-1111-1111-111111111111"]
  }
}
