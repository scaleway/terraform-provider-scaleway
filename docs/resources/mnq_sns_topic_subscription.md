---
subcategory: "Messaging and Queuing"
page_title: "Scaleway: scaleway_mnq_sns_topic_subscription"
---

# Resource: scaleway_mnq_sns_topic_subscription

Manages Scaleway Messaging and Queuing SNS topic subscriptions.
For further information, see
our [main documentation](https://www.scaleway.com/en/docs/messaging/reference-content/sns-overview/).

## Example Usage

### Basic

```terraform
// For default project in default region
resource "scaleway_mnq_sns" "main" {}

resource scaleway_mnq_sns_credentials main {
  project_id = scaleway_mnq_sns.main.project_id
  permissions {
    can_manage = true
    can_publish = true
    can_receive = true
  }
}

resource "scaleway_mnq_sns_topic" "topic" {
  project_id = scaleway_mnq_sns.main.project_id
  name = "my-topic"
  access_key = scaleway_mnq_sns_credentials.main.access_key
  secret_key = scaleway_mnq_sns_credentials.main.secret_key
}

resource scaleway_mnq_sns_topic_subscription main {
  project_id = scaleway_mnq_sns.main.project_id
  access_key = scaleway_mnq_sns_credentials.main.access_key
  secret_key = scaleway_mnq_sns_credentials.main.secret_key
  topic_id = scaleway_mnq_sns_topic.topic.id
  protocol = "http"
  endpoint = "http://example.com"
}
```

## Argument Reference

The following arguments are supported:


- `protocol` - (Required) Protocol of the SNS topic subscription.

- `topic_id` - (Optional) The ID of the topic. Either `topic_id` or `topic_arn` is required. Conflicts with `topic_arn`.

- `topic_arn` - (Optional) The ARN of the topic. Either `topic_id` or `topic_arn` is required.

- `access_key` - (Optional) The access key of the SNS credentials.

- `secret_key` - (Optional) The secret key of the SNS credentials.

- `redrive_policy` - (Optional) Activate JSON redrive policy.

- `sns_endpoint` - (Optional) The endpoint of the SNS service. Can contain a {region} placeholder. Defaults to `https://sns.mnq.{region}.scaleway.com`.

- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`). The [region](../guides/regions_and_zones.md#regions)
  in which SNS is enabled.

- `project_id` - (Defaults to [provider](../index.md#arguments-reference) `project_id`) The ID of the Project in which SNS is enabled.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the topic with format `{region}/{project-id}/{topic-name}/{subscription-id}`

- `arn` - The ARN of the topic subscription

## Import

SNS topic subscriptions can be imported using `{region}/{project-id}/{topic-name}/{subscription-id}`, e.g.

```bash
terraform import scaleway_mnq_sns_topic_subscription.main fr-par/11111111111111111111111111111111/my-topic/11111111111111111111111111111111
```
