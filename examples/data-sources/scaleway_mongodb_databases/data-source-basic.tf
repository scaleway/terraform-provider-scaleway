resource "scaleway_mongodb_instance" "main" {
  name        = "foobar"
  version     = "7.0"
  node_type   = "MGDB-PLAY2-NANO"
  node_number = 1
  user_name   = "my_initial_user"
  password    = "thiZ_is_v&ry_s3cret"
}

data "scaleway_mongodb_databases" "db" {
  instance_id = scaleway_mongodb_instance.main.id
  region      = "fr-par"
}

output "database_names" {
  value = [for database in data.scaleway_mongodb_databases.db.databases : database.name]
}
