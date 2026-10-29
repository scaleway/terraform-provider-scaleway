# Retrieve a VPC ingress rule by its ID
data "scaleway_vpc_ingress_rule" "by_id" {
  ingress_rule_id = "fr-par/11111111-1111-1111-1111-111111111111"
}
