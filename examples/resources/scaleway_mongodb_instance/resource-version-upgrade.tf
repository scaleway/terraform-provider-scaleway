### Example Version Upgrade

# Initial creation with MongoDB 7.0
resource "scaleway_mongodb_instance" "main" {
  name              = "my-mongodb"
  version           = "7.0"
  node_type         = "MGDB-PLAY2-NANO"
  node_number       = 1
  user_name         = "my_initial_user"
  password          = "thiZ_is_v&ry_s3cret"
  volume_size_in_gb = 5
}

# To upgrade to MongoDB 8.0, simply change the version value
# This may trigger a blue/green upgrade that updates the Terraform state with a new instance ID
# resource "scaleway_mongodb_instance" "main" {
#   name              = "my-mongodb"
#   version           = "8.0" # Changed from 7.0
#   node_type         = "MGDB-PLAY2-NANO"
#   node_number       = 1
#   user_name         = "my_initial_user"
#   password          = "thiZ_is_v&ry_s3cret"
#   volume_size_in_gb = 5
# }
