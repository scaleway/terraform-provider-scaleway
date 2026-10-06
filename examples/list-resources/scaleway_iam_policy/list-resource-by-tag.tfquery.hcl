// List policies by tag
list "scaleway_iam_policy" "by_tag" {
  provider = scaleway

  config {
    organization_id = "11111111-1111-1111-1111-111111111111"
    tag             = "production"
  }
}
