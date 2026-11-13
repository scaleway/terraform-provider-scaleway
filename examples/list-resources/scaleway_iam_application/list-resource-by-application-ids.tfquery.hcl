# List applications filtered by application IDs
list "scaleway_iam_application" "by_application_ids" {
  provider = scaleway

  config {
    application_ids = [
      "11111111-1111-1111-1111-111111111111",
      "22222222-2222-2222-2222-222222222222",
    ]
  }
}
