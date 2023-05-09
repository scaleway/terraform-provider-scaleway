---
subcategory: "Messaging and Queuing"
page_title: "Scaleway: scaleway_mnq_queue"
---

# scaleway_mnq_queue

Creates and manages Scaleway Messaging and Queuing queues.

For more information about MNQ, see [the documentation](https://www.scaleway.com/en/developers/api/messaging-and-queuing/).

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
  name         = "my-queue"
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
  name         = "my-queue"

  sqs {
    access_key = scaleway_mnq_credential.main.sqs_sns_credentials.0.access_key
    secret_key = scaleway_mnq_credential.main.sqs_sns_credentials.0.secret_key
  }
}
```

### Argument Reference

The following arguments are supported:

* `namespace_id` - (Required) The ID of the Namespace associated to.

* `name` - (Optional) The name of the queue. Either `name` or `name_prefix` is required. Conflicts with `name_prefix`.

* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.

* `fifo_queue` - (Optional) Specifies whether to create a FIFO queue.

* `message_max_age` - (Optional) The number of seconds the queue retains a message. Must be between 60 and 1_209_600. Defaults to 345_600.

* `message_max_size` - (Optional) The maximum size of a message. Should be in bytes. Must be between 1024 and 262_144. Defaults to 262_144.

* `sqs` - (Optional) The SQS attributes of the queue. Conflicts with `nats`.
  ~ `endpoint` - (Optional) The endpoint of the SQS queue. Can contain a {region} placeholder. Defaults to `http://sqs-sns.mnq.{region}.scw.cloud`.
  ~ `access_key` - (Required) The access key of the SQS queue.
  ~ `secret_key` - (Required) The secret key of the SQS queue.
  ~ `content_based_deduplication` - (Optional) Specifies whether to enable content-based deduplication. Defaults to `false`.
  ~ `receive_wait_time_seconds` - (Optional) The number of seconds to wait for a message to arrive in the queue before returning. Must be between 0 and 20. Defaults to 0.
  ~ `visibility_timeout_seconds` - (Optional) The number of seconds a message is hidden from other consumers. Must be between 0 and 43_200. Defaults to 30.

* `nats` - (Optional) The NATS attributes of the queue. Conflicts with `sqs`.
  ~ `endpoint` - (Optional) The endpoint of the NATS queue. Can contain a {region} placeholder. Defaults to `nats://nats.mnq.{region}.scw.cloud:4222`.
  ~ `credentials` - (Required) Line jump separated key and seed.


### Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `sqs` - The SQS attributes of the queue.
  ~ `url` - The URL of the queue.

### Import

Queues can be imported using the `{region}/{namespace-id}/{queue-name}` format:

```shell
$ terraform import scaleway_mnq_queue.my_queue fr-par/11111111111111111111111111111111/my-queue
```
