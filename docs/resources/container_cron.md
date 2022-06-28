---
page_title: "Scaleway: scaleway_container_cron"
description: |-
Manages Scaleway Containers Triggers.
---

# scaleway_container_cron

Creates and manages Scaleway Container Triggers. For the moment, the feature is limited to CRON Schedule (time-based).

For more information consult
the [documentation](https://www.scaleway.com/en/docs/compute/containers/api-cli/cont-uploading-with-serverless-framework/#configuring-events)
.

For more details about the limitation
check [containers-limitations](https://www.scaleway.com/en/docs/compute/containers/reference-content/containers-limitations/)
.

You can check also
our [containers cron api documentation](https://developers.scaleway.com/en/products/containers/api/#crons-942bf4).

## Example Usage

```hcl
resource scaleway_container_namespace main {
}

resource scaleway_container main {
    name = "my-container-with-cron-tf"
    namespace_id = scaleway_container_namespace.main.id
}

resource scaleway_container_cron main {
    container_id = scaleway_container.main.id
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

## Arguments Reference

The following arguments are required:

- `schedule` - (Required) Cron format string, e.g. @hourly, as schedule time of its jobs to be created and
  executed.
- `container_id` - (Required) The container ID to link with your cron.
- `args`   - (Required) The key-value mapping to define arguments that will be passed to your container’s event object
  during

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions)
  in where the job was created.
- `status` - The cron status.

## Import

Container Cron can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_container_cron.main fr-par/11111111-1111-1111-1111-111111111111
```
