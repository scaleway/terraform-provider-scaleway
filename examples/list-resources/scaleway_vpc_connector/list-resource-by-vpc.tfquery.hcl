# List VPC connectors attached to a specific source VPC
list "scaleway_vpc_connector" "by_vpc" {
  provider = scaleway

  config {
    regions = ["fr-par"]
    vpc_id  = "11111111-1111-1111-1111-111111111111"
  }
}
