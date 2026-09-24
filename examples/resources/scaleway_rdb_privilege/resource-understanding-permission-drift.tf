### Understanding Permission Drift

resource "scaleway_rdb_privilege" "app" {
  instance_id   = scaleway_rdb_instance.main.id
  user_name     = "app_user"
  database_name = "mydb"
  permission    = "readwrite"

  # Later, after new objects are created externally:
  # effective_permission = "custom"  (computed)
  # permission_status    = "drifted" (computed)
}
