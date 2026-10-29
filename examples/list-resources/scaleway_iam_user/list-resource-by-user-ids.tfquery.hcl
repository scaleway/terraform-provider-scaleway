// List users filtered by user IDs
list "scaleway_iam_user" "by_user_ids" {
  provider = scaleway

  config {
    user_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
  }
}
