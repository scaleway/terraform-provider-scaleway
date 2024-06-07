---
subcategory: "Messaging and Queuing"
page_title: "Scaleway: scaleway_mnq_sqs"
---

# Resource: scaleway_mnq_sqs

Activate Scaleway Messaging and Queuing SQS in a Project.
For further information, see
our [main documentation](https://www.scaleway.com/en/docs/serverless/messaging/reference-content/sqs-overview/).

## Example Usage

### Basic

Activate SQS in the default Project

```terraform
resource "scaleway_mnq_sqs" "main" {}
```

Activate SQS for a specific project

```terraform
data scaleway_account_project project {
  name = "default"
}

resource "scaleway_mnq_sqs" "for_project" {
  project_id = data.scaleway_account_project.project.id
}
```

## Argument Reference

The following arguments are supported:


- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`). The [region](../guides/regions_and_zones.md#regions)
  in which SQS will be enabled.

- `project_id` - (Defaults to [provider](../index.md#arguments-reference) `project_id`) The ID of the Project in which SQS will be enabled.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Project

~> **Important:** Messaging and Queueing SQS IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `endpoint` - The endpoint of the SQS service for this Project.

## Import

SQS status can be imported using the `{region}/{project_id}`, e.g.

```bash
$ terraform import scaleway_mnq_sqs.main fr-par/11111111111111111111111111111111
```
