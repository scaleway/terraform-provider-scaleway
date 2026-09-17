# List groups filtered by name
list "scaleway_iam_group" "by_name" {
  provider = scaleway

  config {
    name = "my-group"
  }
}
