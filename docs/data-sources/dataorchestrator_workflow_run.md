---
subcategory: "Data Orchestrator"
page_title: "Scaleway: scaleway_dataorchestrator_workflow_run"
---

# Data Source: scaleway_dataorchestrator_workflow_run

Gets information about a Scaleway Data Orchestrator workflow run.

## Example Usage

```terraform
data "scaleway_dataorchestrator_workflow_run" "by_id" {
  run_id = "fr-par/11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `run_id` - (Required) UUID of the workflow run. Can be a regional ID (`{region}/{id}`).
- `region` - (Optional) Region of the workflow run. Derived from a regional ID or the provider default when unset.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the workflow run (`{region}/{id}`).
- `workflow_definition_id` - ID of the workflow definition.
- `workflow_version_id` - ID of the workflow version that was run.
- `status` - Status of the workflow run.
- `trigger_type` - Trigger type of the workflow run.
- `workflow_dsl` - DSL snapshot used for this run.
- `output` - Output of the workflow run, if any.
- `error_message` - Error message of the workflow run, if any.
- `project_id` - Project ID.
- `srn` - Scaleway Resource Name.
- `created_at` - Creation timestamp.
- `updated_at` - Last update timestamp.
- `terminated_at` - Termination timestamp, if any.
