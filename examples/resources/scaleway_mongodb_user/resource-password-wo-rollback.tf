### Create user with Write Only password (not stored in state), update and rollback the password while ensuring the password is not stored in the state

## Generate an ephemeral password (not stored in the state)
ephemeral "random_password" "main" {
  length      = 20
  special     = true
  upper       = true
  lower       = true
  numeric     = true
  min_upper   = 1
  min_lower   = 1
  min_numeric = 1
  min_special = 1
  # Exclude characters that might cause issues in some contexts
  override_special = "!@#$%^&*()_+-=[]{}|;:,.<>?"
}

resource "scaleway_secret" "main" {
  name        = "mongodb-user-password"
  description = "Password for MongoDB user"
}

# Store the generated password in a Write Only data (not stored in the state)
resource "scaleway_secret_version" "main" {
  secret_id       = scaleway_secret.main.id
  data_wo         = ephemeral.random_password.main.result
  data_wo_version = 1
}

resource "scaleway_mongodb_instance" "main" {
  name              = "test-mongodb-user"
  version           = "7.0.12"
  node_type         = "MGDB-PLAY2-NANO"
  node_number       = 1
  user_name         = "initial_user"
  password          = "initial_password123"
  volume_size_in_gb = 5
}

# Create a user, using the ephemeral password in the Write Only password attribute (not stored in the state)
resource "scaleway_mongodb_user" "main" {
  instance_id         = scaleway_mongodb_instance.main.id
  name                = "test_user"
  password_wo         = ephemeral.random_password.main.result
  password_wo_version = scaleway_secret_version.main.data_wo_version

  roles {
    role          = "read_write"
    database_name = "test_db"
  }
}

## Generate a new ephemeral password (not stored in the state)
ephemeral "random_password" "renewed" {
  length      = 20
  special     = true
  upper       = true
  lower       = true
  numeric     = true
  min_upper   = 1
  min_lower   = 1
  min_numeric = 1
  min_special = 1
  # Exclude characters that might cause issues in some contexts
  override_special = "!@#$%^&*()_+-=[]{}|;:,.<>?"
}

# Store the renewed generated password in a Write Only data (not stored in the state)
resource "scaleway_secret_version" "renewed" {
  secret_id       = scaleway_secret.main.id
  data_wo         = ephemeral.random_password.renewed.result
  data_wo_version = 2
}

# Renew the user password
# resource "scaleway_mongodb_user" "main" {
#   instance_id = scaleway_mongodb_instance.main.id
#   name        = "test_user"
#   password_wo = ephemeral.random_password.renewed.result
#   password_wo_version = scaleway_secret_version.renewed.data_wo_version

#   roles {
#     role          = "read_write"
#     database_name = "test_db"
#   }
# }

# Query the first password version as an Ephemeral Resource (not stored in the state)
# ephemeral "secret_version" "main" {
#   secret_id = scaleway_secret.main.id
#   version   = 1
# }


# Rollback the user password to the first version
# resource "scaleway_mongodb_user" "main" {
#   instance_id = scaleway_mongodb_instance.main.id
#   name        = "test_user"
#   password_wo = ephemeral.secret_version.main.data
#   password_wo_version = 1
#   roles {
#     role          = "read_write"
#     database_name = "test_db"
#   }
# }
