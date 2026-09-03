### Basic

# Get info by SSH key name
data "scaleway_iam_ssh_key" "my_key" {
  name = "my-key-name"
}

# Get info by SSH key id
data "scaleway_iam_ssh_key" "my_key" {
  ssh_key_id = "11111111-1111-1111-1111-111111111111"
}
