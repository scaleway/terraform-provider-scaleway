# List groups filtered by tag
list "scaleway_iam_group" "by_tag" {
  provider = scaleway

  config {
    tag = "production"
  }
}
