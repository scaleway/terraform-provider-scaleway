---
subcategory: "Messaging and Queuing"
page_title: "Scaleway: scaleway_mnq_credential"
description: |-
Manages Scaleway Messaging and Queuing Credential.
---

# scaleway_mnq_credential

This Terraform configuration creates and manage a Scaleway MNQ credential associated with a namespace.
For additional details, kindly refer to our [website](https://www.scaleway.com/en/docs/serverless/messaging/) and
the [API documentation](https://developers.scaleway.com/en/products/messaging_and_queuing/api/v1alpha1/#post-67608e)

## Examples

### NATS credential

```hcl
resource "scaleway_mnq_namespace" "main" {
  name     = "mnq-ns"
  protocol = "nats"
}

resource "scaleway_mnq_credential" "main" {
  name         = "creed-ns"
  namespace_id = scaleway_mnq_namespace.main.id
}
```

### SNS credential

```hcl
resource "scaleway_mnq_namespace" "main" {
  name     = "your-namespace"
  protocol = "sqs_sns"
}

resource "scaleway_mnq_credential" "main" {
  name         = "your-creed-sns"
  namespace_id = scaleway_mnq_namespace.main.id
  sqs_sns_credentials {
    permissions {
      can_publish = true
      can_receive = true
      can_manage  = true
    }
  }
}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Optional) The credential name..
- `namespace_id` - (Required) The namespace containing the Credential.
- `sqs_sns_credentials` - Credential used to connect to the SQS/SNS service.
    - `permissions` List of permissions associated to this Credential. Only one of permissions may be set.
        - `can_publish` - (Optional). Defines if user can publish messages to the service.
        - `can_receive` - (Optional). Defines if user can receive messages from the service.
        - `can_manage` - (Optional). Defines if user can manage the associated resource(s).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The credential ID.
- `protocol` - The protocol associated to the Credential. Possible values are `nats` and `sqs_sns`.
- `sqs_sns_credentials` - The credential used to connect to the SQS/SNS service.
    - `access_key` - The ID of the key.
    - `secret_key` - The Secret value of the key.
- `nats_credentials` - Credentials file used to connect to the NATS service.
    - `content` - Raw content of the NATS credentials file.
- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions)
  in which the namespace should be created.

~> **Important:** Messaging and Queuing credentials' IDs are [regional](../guides/regions_and_zones.md#resource-ids),
which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

## Import

Credential can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_mnq_credential.main fr-par/11111111111111111111111111111111
```
