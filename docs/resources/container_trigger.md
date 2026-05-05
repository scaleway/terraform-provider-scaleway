---
subcategory: "Containers"
page_title: "Scaleway: scaleway_container_trigger"
---

# Resource: scaleway_container_trigger

The `scaleway_container_trigger` resource allows you to create and manage triggers for Scaleway [Serverless Containers](https://www.scaleway.com/en/docs/serverless/containers/).

Refer to the Containers triggers [documentation](https://www.scaleway.com/en/docs/serverless/containers/how-to/add-trigger-to-a-container/) and [API documentation](https://www.scaleway.com/en/developers/api/serverless-containers/#path-triggers-list-all-triggers) for more information.

## Example Usage

### SQS

```terraform
resource "scaleway_container_trigger" "main" {
  container_id = scaleway_container.main.id
  name         = "my-sqs-trigger"
  destination_config {
    http_path = "/"
    http_method = "get"
  }
  sqs {
    endpoint = scaleway_mnq_sqs_queue.main.sqs_endpoint
    queue_url = scaleway_mnq_sqs_queue.main.url
    access_key = scaleway_mnq_sqs_credentials.main.access_key
    secret_key = scaleway_mnq_sqs_credentials.main.secret_key
    # If region is different
    region = scaleway_mnq_sqs.main.region
  }
}
```

### NATS

```terraform
resource "scaleway_container_trigger" "main" {
  container_id = scaleway_container.main.id
  name         = "my-nats-trigger"
  destination_config {
    http_path = "/ping"
    http_method = "get"
  }
  nats {
    subject = "TestSubject"
    server_urls = [ scaleway_mnq_nats_account.main.endpoint ]
    credentials_file_content = scaleway_mnq_nats_credentials.main.file
    # If region is different
    region = scaleway_mnq_nats_account.main.region
  }
}
```

### Cron

```terraform
resource "scaleway_container_trigger" "main" {
  container_id = scaleway_container.main.id
  name         = "my-cron-trigger"
  destination_config {
    http_path = "/patch/here"
    http_method = "patch"
  }
  cron {
    schedule = "5 4 1 * *" #cron at 04:05 on day-of-month 1
    timezone = "Europe/Paris"
    body = "{\"message\": \"This is the content to send to the container.\"}"
    headers = {
      Content-Length = 45
      Content-Type = "application/json"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

- `container_id` - (Required) The unique identifier of the container to create a trigger for.

~> **Important:** Updates to this field will recreate the resource.

- `name` - (Optional) The unique name of the trigger. If not provided, a random name is generated.

- `description` - (Optional) The description of the trigger.

- `tags` - (Optional) The list of tags associated with the trigger.

- `destination_config` - (Required) The configuration of the destination to trigger.
    - `http_path` - (Required) The HTTP path to send the request to (e.g., "/my-webhook-endpoint").
    - `http_method` - (Required) The HTTP method to use when sending the request (e.g., get, post, put, patch, delete).

- `sqs` - The configuration of the Scaleway SQS queue used by the trigger
    - `endpoint` - (Required) Endpoint URL to use to access SQS (e.g., "https://sqs.mnq.fr-par.scaleway.com").
    - `queue_url` - (Required) The URL of the SQS queue to monitor for messages.
    - `access_key` - (Required, Sensitive) The access key for accessing the SQS queue.
    - `secret_key` - (Required, Sensitive) The secret key for accessing the SQS queue.
    - `project_id` - (Optional) The ID of the project in which SQS is enabled, (defaults to [provider](../index.md#arguments-reference) `project_id`)
    - `region` - (Optional) Region where SQS is enabled (defaults to [provider](../index.md#arguments-reference) `region`)
    - `queue` - (Deprecated) The name of the SQS queue.  This argument is no longer supported.

- `nats` - The configuration for the Scaleway NATS account used by the trigger
    - `subject` - (Required) NATS subject to subscribe to (e.g., \"my-subject\")."
    - `server_urls` - (Required) The list of URLs of the NATS server (e.g., "nats://nats.mnq.fr-par.scaleway.com:4222").
    - `credentials_file_content` - (Required, Sensitive) The content of the NATS credentials file that will be used to authenticate with the NATS server and subscribe to the specified subject.
    - `project_id` - (Optional) The ID of the project that contains the Messaging and Queuing NATS account (defaults to [provider](../index.md#arguments-reference) `project_id`)
    - `region` - (Optional) Region where the Messaging and Queuing NATS account is enabled (defaults to [provider](../index.md#arguments-reference) `region`)
    - `account_id` - (Deprecated) unique identifier of the Messaging and Queuing NATS account  .

- `cron` - The configuration for the cron source of the trigger
    - `schedule` - (Required) UNIX cron schedule to run job (e.g., "* * * * *").
    - `timezone` - (Required) Timezone for the cron schedule, in tz database format (e.g., "Europe/Paris").
    - `body` - (Optional) Body to send to the container when the trigger is invoked.
    - `headers` - (Optional) Additional headers to send to the container when the trigger is invoked.

- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`). The [region](../guides/regions_and_zones.md#regions) in which the namespace is created.

## Attributes Reference

The `scaleway_container_trigger` resource exports certain attributes once the Container trigger is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

- `id` - The unique identifier of the Container trigger

~> **Important:** Container trigger IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`.

## Import

Container Triggers can be imported using `{region}/{id}`, as shown below:

```bash
terraform import scaleway_container_trigger.main fr-par/11111111-1111-1111-1111-111111111111
```
