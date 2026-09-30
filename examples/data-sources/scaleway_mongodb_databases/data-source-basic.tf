## Example Usage

resource "scaleway_mongodb_instance" "main" {
  name        = "test-mongodb-databases-datasource"
  version     = "7.0"
  node_type   = "MGDB-PLAY2-NANO"
  node_number = 1
  user_name   = "my_initial_user"
  password    = "thiZ_is_v&ry_s3cret"
}

data "scaleway_mongodb_databases" "main" {
  instance_id = scaleway_mongodb_instance.main.id
}
