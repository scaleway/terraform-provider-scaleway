---
subcategory: "Functions"
page_title: "Scaleway: scaleway_function_trigger"
---

# Resource: scaleway_function_trigger

The `scaleway_function_trigger` resource allows you to create and manage triggers for Scaleway [Serverless Functions](https://www.scaleway.com/en/docs/serverless/functions/).

Refer to the Functions triggers [documentation](https://www.scaleway.com/en/docs/serverless/functions/how-to/add-trigger-to-a-function/) and [API documentation](https://www.scaleway.com/en/developers/api/serverless-functions/#path-triggers-list-all-triggers) for more information.

## Example Usage

### SQS

```terraform
resource scaleway_function_trigger main {
  function_id = scaleway_function.main.id
  name = "my-trigger"
  sqs {
    project_id = scaleway_mnq_sqs.main.project_id
    queue = "MyQueue"
    # If region is different
    region = scaleway_mnq_sqs.main.region
  }
}
```

### NATS

```terraform
resource scaleway_function_trigger main {
  function_id = scaleway_function.main.id
  name = "my-trigger"
  nats {
    account_id = scaleway_mnq_nats_account.main.id
    subject = "MySubject"
    # If region is different
    region = scaleway_mnq_nats_account.main.region
  }
}
```

## Argument Reference

The following arguments are supported:

- `function_id` (Required) The unique identifier of the function to create a trigger for.

- `name` - (Optional) The unique name of the trigger. If not provided, a random name is generated.

- `description` (Optional) The description of the trigger.

- `sqs` The configuration of the Scaleway SQS queue used by the trigger
    - `namespace_id` (Deprecated) ID of the Messaging and Queuing namespace. This argument is deprecated.
    - `queue` (Required) The name of the SQS queue.
    - `project_id` (Optional) The ID of the project in which SQS is enabled, (defaults to [provider](../index.md#project_id) `project_id`)
    - `region` (Optional) Region where SQS is enabled (defaults to [provider](../index.md#project_id) `region`)

- `nats` The configuration for the Scaleway NATS account used by the trigger
    - `account_id` (Required) unique identifier of the Messaging and Queuing NATS account.
    - `subject` (Required) The subject to listen to.
    - `project_id` (Optional) THe ID of the project that contains the Messaging and Queuing NATS account (defaults to [provider](../index.md#project_id) `project_id`)
    - `region` (Optional) Region where the Messaging and Queuing NATS account is enabled (defaults to [provider](../index.md#project_id) `region`)

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which the namespace is created.

## Attributes Reference

The `scaleway_function_trigger` resource exports certain attributes once the Function trigger is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

- `id` - The unique identifier of the Function trigger

~> **Important:** Function trigger IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`.

## Import

Function Triggers can be imported using `{region}/{id}`, as shown below:

```bash
terraform import scaleway_function_trigger.main fr-par/11111111-1111-1111-1111-111111111111
```
