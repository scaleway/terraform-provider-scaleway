### Creating a Redis cluster using a Write Only password (not stored in state)

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

resource "scaleway_redis_cluster" "password_wo_cluster" {
  name                = "test_redis_password_wo"
  version             = "6.2.7"
  node_type           = "RED1-MICRO"
  user_name           = "my_initial_user"
  password_wo         = ephemeral.random_password.db_password.result
  password_wo_version = 1
  cluster_size        = 1
  tls_enabled         = "true"
}
