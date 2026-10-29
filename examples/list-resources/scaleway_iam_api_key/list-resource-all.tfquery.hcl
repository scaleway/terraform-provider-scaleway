// List all API keys in an organization
list "scaleway_iam_api_key" "all" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
  }
}
