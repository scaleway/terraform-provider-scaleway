resource "scaleway_instance_server" "main" {
  name  = "test-terraform-action-server-basic"
  type  = "DEV1-S"
  image = "ubuntu_jammy"

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.scaleway_instance_server_action.main]
    }
  }
}

action "scaleway_instance_server_action" "main" {
  config {
    action    = "reboot"
    server_id = scaleway_instance_server.main.id
  }
}
