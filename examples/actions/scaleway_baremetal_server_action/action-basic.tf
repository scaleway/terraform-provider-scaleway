### Perform actions on a baremetal server (stop, start, reboot)

resource "scaleway_baremetal_server" "base" {
  name        = "my-baremetal-server"
  description = "Server for action examples"
  offer       = "EM-D2124"
  os          = "ubuntu_jammy"

  lifecycle {
    action_trigger {
      events = [after_update]
      actions = [
        action.scaleway_baremetal_server_action.base_stop,
        action.scaleway_baremetal_server_action.base_start,
        action.scaleway_baremetal_server_action.base_reboot,
      ]
    }
  }
}

action "scaleway_baremetal_server_action" "base_stop" {
  config {
    action    = "stop"
    server_id = scaleway_baremetal_server.base.id
    wait      = true
  }
}

action "scaleway_baremetal_server_action" "base_start" {
  config {
    action    = "start"
    server_id = scaleway_baremetal_server.base.id
    wait      = true
  }
}

action "scaleway_baremetal_server_action" "base_reboot" {
  config {
    action    = "reboot"
    server_id = scaleway_baremetal_server.base.id
    boot_type = "normal"
    wait      = true
  }
}
