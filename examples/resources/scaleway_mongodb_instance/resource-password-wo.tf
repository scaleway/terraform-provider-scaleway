### Creating a MongoDB instance using a Write Only password (not stored in state)

## Generate an ephemeral password (not stored in the state)
ephemeral "random_password" "db_password" {
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

resource "scaleway_mongodb_instance" "password_wo_instance" {
  name                = "test-mongodb-password-wo"
  version             = "7.0.12"
  node_type           = "MGDB-PLAY2-NANO"
  node_number         = 1
  user_name           = "my_initial_user"
  password_wo         = ephemeral.random_password.db_password.result
  password_wo_version = 1
}
