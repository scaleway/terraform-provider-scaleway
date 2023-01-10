---
page_title: "Scaleway: scaleway_mnq_credential"
description: |-
Manages Scaleway Messaging and Queuing Credential.
---

# scaleway_mnq_namespace

This Terraform configuration creates and manage a Scaleway MNQ credential associated with a namespace.
For additional details, kindly refer to our [website](https://www.scaleway.com/en/docs/serverless/messaging/)
and the [godoc](https://pkg.go.dev/github.com/scaleway/scaleway-sdk-go@master/api/mnq/v1alpha1#pkg-index).

## Examples

### NATS credential

```hcl
resource "scaleway_mnq_namespace" "main" {
  name     = "mnq-ns"
  protocol = "nats"
  region   = "fr-par"
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

- `name` - (Optional) The name of the namespace.
- `namespace_id` - (Required) it is used to set the ID of the Namespace associated to the credential.
- `sqs_sns_credentials` - The credential used to connect to the SQS/SNS service.
    - `permissions` This field contain permissions which consist of `can_publish`, `can_receive`
      , `can_manage` where the default values are `false`, which are used to determine the permissions associated with
      this Credential.
        - `can_publish` - (Optional). Defines if user can publish messages to the service.
        - `can_receive` - (Optional). Defines if user can receive messages from the service.
        - `can_manage` - (Optional). Defines if user can manage the associated resource(s).

~> **Important:** The `sqs_sns_credentials` and `nats_credentials` field are mutually exclusive, and it can only have
one of the fields, which means only one of the protocol will be used by the namespace.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the namespace
- `protocol` - The protocol of your namespace.
- `nats_credentials` - The credential for NATS protocol.
    - `content` - The content of the NATS credential.
- `sqs_sns_credentials` - The credential used to connect to the SQS/SNS service.
    - `access_key` - The key of the credential.
    - `secret_key` - The secret value of the key.
- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions)
  in which the namespace should be created.

## Import

Credential can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_mnq_credential.main fr-par/11111111111111111111111111111111
```
