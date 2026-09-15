resource "scaleway_mongodb_instance" "main" {
  name        = "test-mongodb-action-snapshot"
  version     = "7.0"
  node_type   = "MGDB-PLAY2-NANO"
  node_number = 1
  user_name   = "my_initial_user"
  password    = "thiZ_is_v&ry_s3cret"

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.scaleway_mongodb_instance_snapshot.main]
    }
  }
}

action "scaleway_mongodb_instance_snapshot" "main" {
  config {
    instance_id = scaleway_mongodb_instance.main.id
    name        = "tf-acc-mongodb-instance-snapshot-action"
    expires_at  = "2026-11-01T00:00:00Z"
    wait        = true
  }
}
