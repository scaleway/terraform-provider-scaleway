### Create a container with Write Only secret environment variables (not stored in state)

resource "scaleway_container_namespace" "main" {
  name        = "my-ns-test"
  description = "test container"
}

resource "scaleway_container" "main" {
  name            = "my-container-wo"
  description     = "write-only secret environment variables test"
  tags            = ["tag1", "tag2"]
  namespace_id    = scaleway_container_namespace.main.id
  registry_image  = "${scaleway_container_namespace.main.registry_endpoint}/alpine:test"
  port            = 9997
  cpu_limit       = 1024
  memory_limit    = 2048
  min_scale       = 3
  max_scale       = 5
  timeout         = 600
  max_concurrency = 80
  privacy         = "private"
  protocol        = "http1"
  deploy          = true

  command = ["bash", "-c", "script.sh"]
  args    = ["some", "args"]

  environment_variables = {
    "foo" = "var"
  }
  secret_environment_variables_wo = {
    "key" = "secret"
  }
  secret_environment_variables_wo_version = 1
}