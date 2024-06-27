---
subcategory: "Messaging and Queuing"
page_title: "Scaleway: scaleway_mnq_sns"
---

# Resource: scaleway_mnq_sns

Activates Scaleway Messaging and Queuing SNS in a Project.
For further information, see
our [main documentation](https://www.scaleway.com/en/docs/serverless/messaging/reference-content/sns-overview/).

## Example Usage

### Basic

Activate SNS in the default Project

```terraform
resource scaleway_mnq_sns "main" {}
```

Activate SNS in a specific Project

```terraform
data scaleway_account_project project {
  name = "default"
}

// For specific Project in default region
resource scaleway_mnq_sns "for_project" {
  project_id = data.scaleway_account_project.project.id
}
```

## Argument Reference

The following arguments are supported:


- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`). The [region](../guides/regions_and_zones.md#regions)
  in which SNS will be enabled.

- `project_id` - (Defaults to [provider](../index.md#arguments-reference) `project_id`) The ID of the project in which SNS will be enabled.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Project

~> **Important:** Messaging and Queueing SN' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `endpoint` - The endpoint of the SNS service for this Project.

## Import

SNS status can be imported using `{region}/{project_id}`, e.g.

```bash
$ terraform import scaleway_mnq_sns.main fr-par/11111111111111111111111111111111
```
