---
subcategory: "Functions"
page_title: "Scaleway: scaleway_function_cron"
---

# Resource: scaleway_function_cron

The `scaleway_function_cron` resource allows you to create and manage CRON triggers for Scaleway [Serverless Functions](https://www.scaleway.com/en/docs/serverless/functions/).

Refer to the Functions CRON triggers [documentation](https://www.scaleway.com/en/docs/serverless/functions/how-to/add-trigger-to-a-function/) and [API documentation](https://www.scaleway.com/en/developers/api/serverless-functions/#path-triggers-list-all-triggers) for more information.

## Example Usage

The following command allows you to add a CRON trigger to a Serverless Function.

```terraform
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
    name = "test-cron"
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

## Argument Reference

The following arguments are supported:

- `schedule` - (Required) CRON format string (refer to the [CRON schedule reference](https://www.scaleway.com/en/docs/serverless/functions/reference-content/cron-schedules/) for more information).

- `function_id` - (Required) The unique identifier of the function to link to your CRON trigger.

- `args` - (Required) The key-value mapping to define arguments that will be passed to your function’s event object

- `name` - (Optional) The name of the function CRON trigger. If not provided, a random name is generated.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions)
  in which the function was created.

## Attributes Reference

The `scaleway_function_cron` resource exports certain attributes once the CRON trigger is retrieved. These attributes can be referenced in other parts of your Terraform configuration.


- `id` - The unique identifier of the function's CRON trigger.

~> **Important:** Function CRON trigger IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions)
  in which the CRON trigger is created.

- `status` - The CRON status.

## Import

Function Cron can be imported using `{region}/{id}`, as shown below:

```bash
terraform import scaleway_function_cron.main fr-par/11111111-1111-1111-1111-111111111111
```
