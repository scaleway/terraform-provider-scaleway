resource "scaleway_job_definition" "main" {
  name                   = "test-jobs-action-start"
  cpu_limit              = 120
  memory_limit           = 256
  local_storage_capacity = 5120
  image_uri              = "docker.io/alpine:latest"
  startup_command        = ["echo", "-e"]
  args                   = ["Hello World"]

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.scaleway_job_definition_start.main]
    }
  }
}

action "scaleway_job_definition_start" "main" {
  config {
    job_definition_id = scaleway_job_definition.main.id
  }
}
