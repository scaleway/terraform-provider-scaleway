# Get info by filters
data "scaleway_vpc_route" "by_filters" {
  vpc_id = scaleway_vpc.my_vpc.id
  tags   = ["my-tag"]
}
