# List applications filtered by tag
list "scaleway_iam_application" "by_tag" {
  provider = scaleway

  config {
    tag = "production"
  }
}
