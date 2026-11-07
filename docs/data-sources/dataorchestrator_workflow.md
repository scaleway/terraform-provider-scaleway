---
subcategory: "Data Orchestrator"
page_title: "Scaleway: scaleway_dataorchestrator_workflow"
---

# Data Source: scaleway_dataorchestrator_workflow

Gets information about a Scaleway Data Orchestrator workflow definition.

## Example Usage

```terraform
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
}

data "scaleway_dataorchestrator_workflow" "by_id" {
  workflow_id = scaleway_dataorchestrator_workflow.main.id
}
```

## Argument Reference

- `workflow_id` - (Required) UUID of the workflow definition. Can be a regional ID (`{region}/{id}`).
- `region` - (Optional) Region of the workflow. Derived from a regional ID or the provider default when unset.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the workflow definition (`{region}/{id}`).
- `name` - Name of the workflow definition.
- `description` - Description of the workflow definition.
- `workflow_dsl` - Current workflow version DSL (YAML).
- `version_name` - Name of the current workflow version.
- `version_description` - Description of the current workflow version.
- `version_id` - ID of the current workflow version.
- `environment_variables` - Environment variables of the workflow.
- `workflow_timeout` - Workflow timeout as a Go duration string.
- `cron_schedule` - Cron schedule of the workflow, if any.
- `project_id` - Project ID.
- `srn` - Scaleway Resource Name.
- `created_at` - Creation timestamp.
- `updated_at` - Last update timestamp.
