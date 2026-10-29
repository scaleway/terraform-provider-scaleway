resource "scaleway_container" "main" {
  name         = "my-container"
  namespace_id = scaleway_container_namespace.main.id
  image        = "nginx:latest"

  liveness_probe {
    http {
      path = "/ping"
    }
    failure_threshold = 40
    interval          = "5s"
    timeout           = "1m"
  }
}
