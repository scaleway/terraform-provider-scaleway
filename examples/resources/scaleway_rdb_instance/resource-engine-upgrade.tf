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

# To upgrade to PostgreSQL 15, change the engine value and explicitly allow the major version upgrade.
# This triggers a blue/green upgrade: read the "Engine upgrade" section of this page before applying.
# resource "scaleway_rdb_instance" "main" {
#   name                        = "my-database"
#   node_type                   = "DB-DEV-S"
#   engine                      = "PostgreSQL-15" # Changed from PostgreSQL-14
#   allow_major_version_upgrade = true
#   is_ha_cluster               = false
#   disable_backup              = true
#   user_name                   = "my_user"
#   password                    = "thiZ_is_v&ry_s3cret"
#
#   timeouts {
#     update = "120m"
#   }
# }
