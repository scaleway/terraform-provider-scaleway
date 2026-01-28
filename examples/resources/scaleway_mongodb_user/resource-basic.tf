### Basic user creation

resource "scaleway_mongodb_instance" "main" {
  name              = "test-mongodb-user"
  version           = "7.0.12"
  node_type         = "MGDB-PLAY2-NANO"
  node_number       = 1
  user_name         = "initial_user"
  password          = "initial_password123"
  volume_size_in_gb = 5
}

resource "scaleway_mongodb_user" "main" {
  instance_id = scaleway_mongodb_instance.main.id
  name        = "my_user"
  password    = "my_password123"

  roles {
    role          = "read_write"
    database_name = "my_database"
  }
}
