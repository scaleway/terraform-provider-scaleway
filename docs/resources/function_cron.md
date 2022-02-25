---
page_title: "Scaleway: scaleway_function_cron"
description: |-
Manages Scaleway Function Cron.
---

# scaleway_function_cron

Creates and manages Scaleway Function Cron.
For more information see [the documentation](https://developers.scaleway.com/en/products/functions/api/).

## Examples

### Basic

```hcl
resource "scaleway_function_cron" "main" {
}
```

## Arguments Reference

The following arguments are supported:


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the cron
- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which the cron should be created.

## Import

Cron can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_function_cron.main fr-par/11111111-1111-1111-1111-111111111111
```
