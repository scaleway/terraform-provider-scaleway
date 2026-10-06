// List API keys filtered by description
list "scaleway_iam_api_key" "by_description" {
  provider = scaleway

  config {
    description = "production"
  }
}
