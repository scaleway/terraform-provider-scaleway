---
subcategory: "Messaging and Queuing"
page_title: "Scaleway: scaleway_mnq_queue"
---

# scaleway_mnq_queue

Creates and manages Scaleway Messaging and Queuing queues.

## Examples

### NATS

```hcl
resource "scaleway_mnq_namespace" "main" {
  protocol = "nats"
}

resource "scaleway_mnq_credential" "main" {
  namespace_id = scaleway_mnq_namespace.main.id
}

resource "scaleway_mnq_queue" "my_queue" {
  namespace_id = scaleway_mnq_namespace.main.id
  name = "my-queue"
}
```

### SQS

```hcl
resource "scaleway_mnq_namespace" "main" {
  protocol = "sqs_sns"
}

resource "scaleway_mnq_credential" "main" {
  namespace_id = scaleway_mnq_namespace.main.id
  sqs_sns_credentials {
    permissions {
      can_publish = true
      can_receive = true
      can_manage  = true
    }
  }
}

resource "scaleway_mnq_queue" "my_queue" {
  namespace_id = scaleway_mnq_namespace.main.id
  name = "my-queue"

  sqs {
    access_key = scaleway_mnq_credential.main.sqs_sns_credentials.0.access_key
    secret_key = scaleway_mnq_credential.main.sqs_sns_credentials.0.secret_key
  }
}
```

### Argument Reference

The following arguments are supported:

* `namespace_id` - (Required) The ID of the Namespace associated to.

* `name` - (Optional) The name of the queue. Either `name` or `name_prefix` is required.

* `name_prefix` - (Optional) The prefix of the queue. Either `name` or `name_prefix` is required.

* `fifo_queue` - (Optional) Whether or not the queue should be a FIFO queue. Defaults to `false`.

* `sqs` - (Optional) The SQS attributes of the queue. See below for details.
  ~ `access_key` - (Required) The access key of the SQS queue.
  ~ `secret_key` - (Required) The secret key of the SQS queue.
  ~ `content_based_deduplication` - (Optional) Whether or not content-based deduplication is enabled for the queue. Defaults to `false`.
  ~ `max_message_size` - (Optional) The maximum size of messages that can be sent to the queue, in bytes. Must be between 1024 and 262144 bytes. Defaults to 2048 bytes.
  ~ `message_retention_seconds` - (Optional) The number of seconds that messages are retained in the queue. Must be between 60 and 1209600 seconds. Defaults to 86400 seconds.
  ~ `receive_wait_time_seconds` - (Optional) The number of seconds that the queue should wait for new messages to arrive before returning an empty response. Defaults to 20 seconds.
  ~ `visibility_timeout_seconds` - (Optional) The number of seconds that a message is hidden from other consumers after it has been received by one consumer. Must be between 0 and 43200 seconds. Defaults to 120 seconds.

### Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `url` - The URL of the queue.

### Import

Queues can be imported using the `{region}/{namespace-id}/{queue-name}` format:

```shell
$ terraform import scaleway_mnq_queue.my_queue fr-par/11111111111111111111111111111111/my-queue
```
