## Example Usage

# Get the database ACL for the instance id 11111111-1111-1111-1111-111111111111 located in the default region e.g: fr-par
data "scaleway_rdb_acl" "my_acl" {
  instance_id = "11111111-1111-1111-1111-111111111111"
}
