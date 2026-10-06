// List API keys filtered by editable status
list "scaleway_iam_api_key" "by_editable" {
  provider = scaleway

  config {
    editable = true
  }
}
