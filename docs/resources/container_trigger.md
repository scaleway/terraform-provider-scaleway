---
subcategory: "Containers"
page_title: "Scaleway: scaleway_container_trigger"
---

# scaleway_container_trigger

Creates and manages Scaleway Container Triggers.
For more information see [the documentation](https://www.scaleway.com/en/developers/api/serverless-containers/#path-triggers).

## Examples

### Basic

```hcl
resource scaleway_container_trigger main {
  container_id = scaleway_container.main.id
  name = "my-trigger"
  sqs {
    project_id = scaleway_mnq_sqs.main.project_id
    queue = "MyQueue"
    # If region is different
    region = scaleway_mnq_sqs.main.region
  }
}
```

## Arguments Reference

The following arguments are supported:

- `container_id` (Required) The ID of the container to create a trigger for

- `name` - (Optional) The unique name of the trigger. Default to a generated name.

- `description` (Optional) The description of the trigger.

- `sqs` The configuration of the Scaleway's SQS used by the trigger
    - `namespace_id` (Optional) ID of the mnq namespace. Deprecated.
    - `queue` (Required) Name of the queue
    - `project_id` (Optional) ID of the project that contain the mnq namespace, defaults to provider's project
    - `region` (Optional) Region where the mnq namespace is, defaults to provider's region

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which the namespace should be created.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the container trigger

- ~> **Important:** Container Triggers' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

## Import

Container Triggers can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_container_trigger.main fr-par/11111111-1111-1111-1111-111111111111
```
