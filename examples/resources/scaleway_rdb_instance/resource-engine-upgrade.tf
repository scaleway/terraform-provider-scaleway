### Example Engine Upgrade

# Initial creation with PostgreSQL 14
resource "scaleway_rdb_instance" "main" {
  name           = "my-database"
  node_type      = "DB-DEV-S"
  engine         = "PostgreSQL-14"
  is_ha_cluster  = false
  disable_backup = true
  user_name      = "my_user"
  password       = "thiZ_is_v&ry_s3cret"
}

# Check available versions for upgrade
output "upgradable_versions" {
  value = scaleway_rdb_instance.main.upgradable_versions
}

# To upgrade to PostgreSQL 15, simply change the engine value
# This will trigger a blue/green upgrade with automatic endpoint migration
# resource "scaleway_rdb_instance" "main" {
#   name           = "my-database"
#   node_type      = "DB-DEV-S"
#   engine         = "PostgreSQL-15"  # Changed from PostgreSQL-14
#   is_ha_cluster  = false
#   disable_backup = true
#   user_name      = "my_user"
#   password       = "thiZ_is_v&ry_s3cret"
# }