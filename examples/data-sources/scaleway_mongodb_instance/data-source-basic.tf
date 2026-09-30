## Example Usage

# Get info by name
data "scaleway_mongodb_instance" "my_instance" {
  name = "foobar"
}

# Get info by instance ID
data "scaleway_mongodb_instance" "my_instance" {
  instance_id = "11111111-1111-1111-1111-111111111111"
}

# Get other attributes
output "mongodb_version" {
  description = "Version of the MongoDB instance"
  value       = data.scaleway_mongodb_instance.my_instance.version
}
