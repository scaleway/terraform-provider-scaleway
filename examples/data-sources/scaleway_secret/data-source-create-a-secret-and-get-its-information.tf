### Create a secret and get its information

// Create a secret
resource "scaleway_secret" "main" {
  name        = "foo"
  description = "barr"
}

// Get the secret information specified by the secret ID
data "scaleway_secret" "my_secret" {
  secret_id = "11111111-1111-1111-1111-111111111111"
}

// Get the secret information specified by the secret name
data "scaleway_secret" "by_name" {
  name = "your_secret_name"
}
