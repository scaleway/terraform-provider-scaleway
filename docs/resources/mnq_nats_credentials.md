---
subcategory: "Messaging and Queuing"
page_title: "Scaleway: scaleway_mnq_nats_credentials"
---

# Resource: scaleway_mnq_nats_credentials

Creates and manages Scaleway Messaging and Queuing NATS credentials.
For further information, see
our [main documentation](https://www.scaleway.com/en/docs/serverless/messaging/reference-content/nats-overview/).

## Example Usage

### Basic

```terraform
resource "scaleway_mnq_nats_account" "main" {
  name = "nats-account"
}

resource "scaleway_mnq_nats_credentials" "main" {
  account_id = scaleway_mnq_nats_account.main.id
}
```

## Argument Reference

The following arguments are supported:

- `account_id` - (Required) The ID of the NATS account the credentials are generated from

- `name` - (Optional) The unique name of the NATS credentials.

- `region` - (Defaults to [provider](../index.mds#arguments-reference) `region`). The [region](../guides/regions_and_zones.md#regions)
  in which the account exists.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the credentials

~> **Important:** Messaging and Queueing NATS credentials' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `file` - The content of the credentials file.

## Import

Namespaces can be imported using `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_mnq_nats_credentials.main fr-par/11111111111111111111111111111111
```
