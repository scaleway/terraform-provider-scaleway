### Multiple user creation

resource "scaleway_mongodb_instance" "main" {
  name              = "test-mongodb-multi-user"
  version           = "7.0.12"
  node_type         = "MGDB-PLAY2-NANO"
  node_number       = 1
  user_name         = "admin_user"
  password          = "admin_password123"
  volume_size_in_gb = 5
}

resource "scaleway_mongodb_user" "app_user" {
  instance_id = scaleway_mongodb_instance.main.id
  name        = "app_user"
  password    = "app_password123"

  roles {
    role          = "read_write"
    database_name = "app_database"
  }

  roles {
    role          = "read"
    database_name = "logs_database"
  }
}

resource "scaleway_mongodb_user" "admin_user" {
  instance_id = scaleway_mongodb_instance.main.id
  name        = "admin_user"
  password    = "admin_password123"

  roles {
    role          = "db_admin"
    database_name = "admin"
  }

  roles {
    role         = "read"
    any_database = true
  }
}
