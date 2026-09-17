# List VPC connectors filtered by name (matches connectors whose name contains "prod")
list "scaleway_vpc_connector" "by_name" {
  provider = scaleway

  config {
    regions = ["*"]
    name    = "prod"
  }
}
