# Get info by route ID
data "scaleway_vpc_route" "by_id" {
  route_id = "fr-par/11111111-1111-1111-1111-111111111111"
}
