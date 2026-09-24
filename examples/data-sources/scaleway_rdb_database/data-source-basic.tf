## Example Usage

# Get the database foobar hosted on instance id 11111111-1111-1111-1111-111111111111
data "scaleway_rdb_database" "my_db" {
  instance_id = "11111111-1111-1111-1111-111111111111"
  name        = "foobar"
}
