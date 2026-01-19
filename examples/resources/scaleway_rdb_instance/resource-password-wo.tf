### Usage of ephemeral random_password for instance password without storing it in state

// Generate an ephemeral password (not stored in the state)
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

// Pass the ephemeral password with password_wo (not stored in the state)
resource "scaleway_rdb_instance" "main" {
  name                = "test-rdb"
  node_type           = "DB-DEV-S"
  engine              = "PostgreSQL-15"
  is_ha_cluster       = true
  disable_backup      = true
  user_name           = "my_initial_user"
  password_wo         = ephemeral.random_password.db_password.result
  password_wo_version = 1
  encryption_at_rest  = true
}