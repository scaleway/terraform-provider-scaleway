### Reboot an Apple Silicon server after creation

resource "scaleway_apple_silicon_server" "main" {
  name             = "my-apple-silicon-server"
  type             = "M4-M"
  public_bandwidth = 1000000000

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.scaleway_apple_silicon_reboot_server.main_reboot]
    }
  }
}

action "scaleway_apple_silicon_reboot_server" "main_reboot" {
  config {
    server_id = scaleway_apple_silicon_server.main.id
    wait      = true
  }
}
