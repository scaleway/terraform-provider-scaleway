# Retrieve a VPC connector by name
data "scaleway_vpc_connector" "by_name" {
  name = "my-vpc-connector"
}
