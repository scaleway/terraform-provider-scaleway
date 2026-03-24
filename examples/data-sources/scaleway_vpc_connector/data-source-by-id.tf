# Retrieve a VPC connector by its ID
data "scaleway_vpc_connector" "by_id" {
  connector_id = "fr-par/11111111-1111-1111-1111-111111111111"
}
