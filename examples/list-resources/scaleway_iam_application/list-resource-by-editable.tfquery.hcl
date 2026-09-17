# List applications filtered by editable status
list "scaleway_iam_application" "by_editable" {
  provider = scaleway

  config {
    editable = true
  }
}
