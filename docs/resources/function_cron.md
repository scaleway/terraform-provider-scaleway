---
page_title: "Scaleway: scaleway_function_cron"
description: |-
Manages Scaleway Functions Triggers.
---

# scaleway_function_cron

Creates and manages Scaleway Function Triggers. For the moment, the feature is limited to CRON Schedule (time-based).

For more information consult
the [documentation](https://www.scaleway.com/en/docs/compute/functions/api-cli/fun-uploading-with-serverless-framework/#configuring-events)
.

For more details about the limitation
check [functions-limitations](https://www.scaleway.com/en/docs/compute/functions/reference-content/functions-limitations/).

You can check also
our [functions cron api documentation](https://developers.scaleway.com/en/products/functions/api/#crons-942bf4).

## Example Usage

```hcl
resource scaleway_function_namespace main {
    name = "test-cron"
}

resource scaleway_function main {
    name = "test-cron"
    namespace_id = scaleway_function_namespace.main.id
    runtime = "node14"
    privacy = "private"
    handler = "handler.handle"
}

resource scaleway_function_cron main {
    function_id = scaleway_function.main.id
    schedule = "0 0 * * *"
    args = jsonencode({test = "scw"})
}

resource scaleway_function_cron func {
    function_id = scaleway_function.main.id
    schedule = "0 1 * * *"
    args = jsonencode({my_var = "terraform"})
}
```

## Arguments Reference

The following arguments are required:

- `schedule` - (Required) Cron format string, e.g. @hourly, as schedule time of its jobs to be created and
  executed.
- `function_id` - (Required) The function ID to link with your cron.
- `args`   - (Required) The key-value mapping to define arguments that will be passed to your function’s event object
  during

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions)
  in where the job was created.
- `status` - The cron status.

## Import

Container Cron can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_function_cron.main fr-par/11111111-1111-1111-1111-111111111111
```
