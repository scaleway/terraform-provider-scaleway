# List applications filtered by name
list "scaleway_iam_application" "by_name" {
  provider = scaleway

  config {
    name = "my-application"
  }
}
