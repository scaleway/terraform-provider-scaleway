### Basic

# Get info by name
data "scaleway_iam_group" "find_by_name" {
  name = "foobar"
}

# Get info by group ID
data "scaleway_iam_group" "find_by_id" {
  group_id = "11111111-1111-1111-1111-111111111111"
}
