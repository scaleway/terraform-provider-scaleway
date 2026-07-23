### Basic

# Get info by name
data "scaleway_iam_application" "find_by_name" {
  name = "foobar"
}
# Get info by application ID
data "scaleway_iam_application" "find_by_id" {
  application_id = "11111111-1111-1111-1111-111111111111"
}
