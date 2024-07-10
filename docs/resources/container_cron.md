---
subcategory: "Containers"
page_title: "Scaleway: scaleway_container_cron"
---

# Resource: scaleway_container_cron

The `scaleway_container_cron` resource allows you to create and manage CRON triggers for Scaleway [Serverless Containers](https://www.scaleway.com/en/docs/serverless/containers/).

Refer to the Containers CRON triggers [documentation](https://www.scaleway.com/en/docs/serverless/containers/how-to/add-trigger-to-a-container/) and [API documentation](https://www.scaleway.com/en/developers/api/serverless-containers/#path-triggers-list-all-triggers) for more information.

## Add a CRON trigger to a container

The following command allows you to add a CRON trigger to a Serverless Container.

```terraform
resource scaleway_container_namespace main {
}

resource scaleway_container main {
    name = "my-container-with-cron-tf"
    namespace_id = scaleway_container_namespace.main.id
}

resource scaleway_container_cron main {
    container_id = scaleway_container.main.id
    name = "my-cron-name"
    schedule = "5 4 1 * *" #cron at 04:05 on day-of-month 1
    args = jsonencode(
    {
        address   = {
            city    = "Paris"
            country = "FR"
        }
        age       = 23
        firstName = "John"
        isAlive   = true
        lastName  = "Smith"
        # minScale: 1
        # memoryLimit: 256
        # maxScale: 2
        # timeout: 20000
        # Local environment variables - used only in given function
    }
    )
}
```

## Argument Reference

The following arguments are supported:

- `schedule` - (Required) CRON format string (refer to the [CRON schedule reference](https://www.scaleway.com/en/docs/serverless/containers/reference-content/cron-schedules/) for more information).

- `container_id` - (Required) The unique identifier of the container to link to your CRON trigger.

- `args` - (Required) The key-value mapping to define arguments that will be passed to your container’s event object

- `name` - (Optional) The name of the container CRON trigger. If not provided, a random name is generated.

## Attributes Reference

The `scaleway_container_cron` resource exports certain attributes once the CRON trigger is retrieved. These attributes can be referenced in other parts of your Terraform configuration.


- `id` - The unique identifier of the container's CRON trigger.

~> **Important:** Container CRON trigger IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions)
  in which the CRON trigger is created.

- `status` - The CRON status.

## Import

Container Cron can be imported using `{region}/{id}`, as shown below:

```bash
terraform import scaleway_container_cron.main fr-par/11111111-1111-1111-1111-111111111111
```
