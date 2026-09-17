resource "scaleway_container_namespace" "main" {}

resource "scaleway_container" "main" {
  name         = "my-container"
  description  = "This container has a description."
  tags         = ["tag1", "tag2"]
  namespace_id = scaleway_container_namespace.main.id
  image        = "nginx:latest"
  port         = 80

  cpu_limit          = 1024
  memory_limit_bytes = 2048000000
  min_scale          = 3
  max_scale          = 5
  timeout            = 600
  protocol           = "http1"

  command = ["bash", "-c", "script.sh"]
  args    = ["some", "args"]

  environment_variables = {
    "foo" = "var"
  }
  secret_environment_variables = {
    "key" = "secret"
  }
}
