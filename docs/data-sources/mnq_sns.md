---
subcategory: "Messaging and Queuing"
page_title: "Scaleway: scaleway_mnq_sns"
---

# scaleway_mnq_sns

Gets information about SNS for a Project

## Examples

### Basic

```hcl
// For default project
data "scaleway_mnq_sns" "main" {}

// For specific project
data "scaleway_mnq_sns" "for_project" {
  project_id = scaleway_account_project.main.id
}
```

## Arguments Reference

The following arguments are supported:

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which sns is enabled.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the Project in which sns is enabled.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Project

~> **Important:** Messaging and Queueing sns' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `endpoint` - The endpoint of the SNS service for this Project.
