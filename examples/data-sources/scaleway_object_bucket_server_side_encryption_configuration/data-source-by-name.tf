# Get by Bucket Name
data "scaleway_object_bucket_server_side_encryption_configuration" "by_bucket" {
  bucket = scaleway_object_bucket.main.name
}
