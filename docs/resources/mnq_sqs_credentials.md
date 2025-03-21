---
subcategory: "Messaging and Queuing"
page_title: "Scaleway: scaleway_mnq_sqs_credentials"
---

# Resource: scaleway_mnq_sqs_credentials

Creates and manages Scaleway Messaging and Queuing SQS credentials.
For further information, see
our [main documentation](https://www.scaleway.com/en/docs/messaging/reference-content/sqs-overview/)

## Example Usage

### Basic

```terraform
resource "scaleway_mnq_sqs" "main" {}

resource scaleway_mnq_sqs_credentials main {
  project_id = scaleway_mnq_sqs.main.project_id
  name = "sqs-credentials"

  permissions {
    can_manage = false
    can_receive = true
    can_publish = false
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Optional) The unique name of the SQS credentials.

- `permissions` - (Optional). List of permissions associated with these credentials. Only one of the following permissions may be set:
    - `can_publish` - (Optional). Defines whether the user can publish messages to the service.
    - `can_receive` - (Optional). Defines whether the user can receive messages from the service.
    - `can_manage` - (Optional). Defines whether the user can manage the associated resource(s).


- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`). The [region](../guides/regions_and_zones.md#regions) in which SQS is enabled.

- `project_id` - (Defaults to [provider](../index.md#arguments-reference) `project_id`) The ID of the Project in which SQS is enabled.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the credentials

~> **Important:** Messaging and Queueing SQS credential IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `access_key` - The ID of the key.
- `secret_key` - The secret value of the key.

## Import

SQS credentials can be imported using `{region}/{id}`, e.g.

```bash
terraform import scaleway_mnq_sqs_credentials.main fr-par/11111111111111111111111111111111
```
