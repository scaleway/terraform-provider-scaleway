---
subcategory: "Messaging and Queuing"
page_title: "Scaleway: scaleway_mnq_nats_account"
---

# scaleway_mnq_nats_account

Creates and manages Scaleway Messaging and queuing Nats Accounts.
For further information please check
our [documentation](https://www.scaleway.com/en/docs/serverless/messaging/reference-content/nats-overview/)

## Examples

### Basic

```hcl
resource "scaleway_mnq_nats_account" "main" {
  name = "nats-account"
}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Optional) The unique name of the nats account.

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions)
  in which the account should be created.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the
  account is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the account

~> **Important:** Messaging and Queueing nats accounts' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `endpoint` - The endpoint of the NATS service for this account.

## Import

Namespaces can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_mnq_namespace.main fr-par/11111111111111111111111111111111
```
