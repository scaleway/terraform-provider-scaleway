// List users filtered by tag
list "scaleway_iam_user" "by_tag" {
  provider = scaleway

  config {
    tag = "production"
  }
}
