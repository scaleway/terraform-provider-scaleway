## Example Usage

data "scaleway_rdb_database_backup" "find_by_name" {
  name = "mybackup"
}

data "scaleway_rdb_database_backup" "find_by_name_and_instance" {
  name        = "mybackup"
  instance_id = "11111111-1111-1111-1111-111111111111"
}

data "scaleway_rdb_database_backup" "find_by_id" {
  backup_id = "11111111-1111-1111-1111-111111111111"
}
