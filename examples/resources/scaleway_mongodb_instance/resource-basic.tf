### Basic MongoDB instance creation

resource "scaleway_mongodb_instance" "main" {
  name              = "test-mongodb-basic1"
  version           = "7.0.12"
  node_type         = "MGDB-PLAY2-NANO"
  node_number       = 1
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  volume_size_in_gb = 5
}
