### Basic

# Get policy by id
data "scaleway_iam_policy" "find_by_id" {
  policy_id = "11111111-1111-1111-1111-111111111111"
}

# Get policy by name
data "scaleway_iam_policy" "find_by_name" {
  name = "my_policy"
}
