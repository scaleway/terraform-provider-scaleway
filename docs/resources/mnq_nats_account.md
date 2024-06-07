---
subcategory: "Messaging and Queuing"
page_title: "Scaleway: scaleway_mnq_nats_account"
---

# Resource: scaleway_mnq_nats_account

Creates and manages Scaleway Messaging and Queuing NATS accounts.
For further information, see
our [main documentation](https://www.scaleway.com/en/docs/serverless/messaging/reference-content/nats-overview/)
To use the Scaleway provider with the official NATS JetStream provider, check out the [corresponding guide](../guides/mnq_with_nats_terraform_provider.md).

## Example Usage

### Basic

```terraform
resource "scaleway_mnq_nats_account" "main" {
  name = "nats-account"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Optional) The unique name of the NATS account.

- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`). The [region](../guides/regions_and_zones.md#regions)
  in which the account should be created.

- `project_id` - (Defaults to [provider](../index.md#arguments-reference) `project_id`) The ID of the Project the
  account is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the account

~> **Important:** Messaging and Queueing NATS account IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `endpoint` - The endpoint of the NATS service for this account.

## Import

Namespaces can be imported using `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_mnq_nats_account.main fr-par/11111111111111111111111111111111
```
