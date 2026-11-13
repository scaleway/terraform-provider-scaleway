# Get by ID
data "scaleway_object_bucket_server_side_encryption_configuration" "by_id" {
  bucket_server_side_encryption_configuration_id = scaleway_object_bucket_server_side_encryption_configuration.main.id
}
