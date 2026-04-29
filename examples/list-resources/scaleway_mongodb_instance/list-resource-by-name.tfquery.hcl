# List MongoDB instances filtered by name prefix
list "scaleway_mongodb_instance" "by_name" {
  provider = scaleway

  config {
    regions     = ["*"]
    name        = "my-mongodb"
  }
}
