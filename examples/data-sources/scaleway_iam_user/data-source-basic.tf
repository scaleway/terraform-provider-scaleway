### Basic

# Get info by user id
data "scaleway_iam_user" "find_by_id" {
  user_id = "11111111-1111-1111-1111-111111111111"
}
# Get info by email address
data "scaleway_iam_user" "find_by_email" {
  email = "foo@bar.com"
}
