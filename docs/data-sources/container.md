---
page_title: "Scaleway: scaleway_container"
description: |-
Gets information about a container.
---

# scaleway_container

Creates and manages Scaleway Container.
For more information consult the [documentation](https://www.scaleway.com/en/docs/faq/serverless-containers/).

For more details about the limitation check [containers-limitations](https://www.scaleway.com/en/docs/compute/containers/reference-content/containers-limitations/).

You can check also our [containers guide](https://www.scaleway.com/en/docs/compute/containers/concepts/).

## Example Usage

```hcl
resource scaleway_container_namespace main {
    name = "%s"
    description = "test container"
}

resource scaleway_container main {
    name = "my-container-02"
    description = "environment variables test"
    namespace_id = scaleway_container_namespace.main.id
    registry_image = "${scaleway_container_namespace.main.endpoint}/alpine:test"
    port = 9997
    cpu_limit = 140
    memory_limit = 256
    min_scale = 3
    max_scale = 5
    timeout = 600
    max_concurrency = 80
    privacy = "private"
    protocol = "h2c"
    redeploy = true

    environment_variables = {
        "foo" = "var"
    }
}
```


## Arguments Reference

The following arguments are supported:

- `name` - (Required) The unique name of the container namespace.

~> **Important** Updates to `name` will recreate the container.

- `description` (Optional) The description of the container.

- `namespace_id` - (Required) The container namespace ID of the container.

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which the namespace should be created.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the namespace is associated with.

- `environment_variables` - The environment variables of the container.

- `min_scale` - The minimum of running container instances continuously. Defaults to 0.

- `max_scale` - The maximum of number of instances this container can scale to. Default to 20.

- `memory_limit` - The memory computing resources in MB to allocate to each container. Defaults to 128.

- `cpu_limit` - The amount of vCPU computing resources to allocate to each container. Defaults  to 70.

- `timeout` - The maximum amount of time in seconds during which your container can process a request before we stop it. Defaults to 300s.

- `privacy` - The privacy type access.

- `registry_image` - The registry image address. e.g: **"rg.fr-par.scw.cloud/$NAMESPACE/$IMAGE"**.

- `max_concurrency` - The maximum the number of simultaneous requests your container can handle at the same time. Defaults to 50.

- `domain_name` - The container domain name.

- `protocol` - The communication protocol. Defaults to http1.

- `port` - The port to expose the container. Defaults to 8080.

- `redeploy` - Allow deploy container.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the container
- `organization_id` - The organization ID the container is associated with.
- `status` - The container status.
- `cron_status` - The cron status of the container.
- `error_message` - The error message of the container.

```bash
$ terraform import scaleway_container.main fr-par/11111111-1111-1111-1111-111111111111
```
