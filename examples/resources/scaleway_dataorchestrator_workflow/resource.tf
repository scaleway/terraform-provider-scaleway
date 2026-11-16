resource "scaleway_dataorchestrator_workflow" "main" {
  name         = "example-workflow"
  description  = "Managed by Terraform"
  region       = "fr-par"
  version_name = "v1-0-0"

  workflow_dsl = <<-EOT
    document:
      dsl: 1.0.3
      namespace: example
      name: example-workflow
      version: 1.0.0
    do:
      - setColor:
          set:
            color: blue
  EOT

  cron_schedule = {
    schedule            = "0 6 * * *"
    timezone_identifier = "Europe/Paris"
    overlap             = "forbid"
  }
}
