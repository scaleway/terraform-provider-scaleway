## Example Usage

# Get the database privilege for the user "my-user" on the database "my-database" hosted on instance id 11111111-1111-1111-1111-111111111111 and on the default region. e.g: fr-par
data "scaleway_rdb_privilege" "main" {
  instance_id   = "11111111-1111-1111-1111-111111111111"
  user_name     = "my-user"
  database_name = "my-database"
}
