---
page_title: "Scaleway: scaleway_container"
description: |-
Gets information about a container.
---

# scaleway_container

Gets information about the Scaleway Container.

For more information consult the [documentation](https://www.scaleway.com/en/docs/faq/serverless-containers/).

For more details about the limitation check [containers-limitations](https://www.scaleway.com/en/docs/compute/containers/reference-content/containers-limitations/).

You can check also our [containers guide](https://www.scaleway.com/en/docs/compute/containers/concepts/).

## Example Usage

```hcl
resource scaleway_container_namespace main {
}

resource scaleway_container main {
    name = "test-container-data"
    namespace_id = scaleway_container_namespace.main.id
}

// Get info by container name
data "scaleway_container" "by_name" {
    namespace_id = scaleway_container_namespace.main.id
    name = scaleway_container.main.name
}

// Get info by container ID
data "scaleway_container" "by_id" {
    namespace_id = scaleway_container_namespace.main.id
    container_id = scaleway_container.main.id
}
```

## Argument Reference

- `name` - (Optional) The container name.
  Only one of `name` and `container_id` should be specified.

- `container_id` - (Optional) The container id.
- `namespace_id` - (Required) The container namespace id
  Only one of `name` and `container_id` should be specified.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the container exists.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the namespace is associated with.


## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `name` - The unique name of the container namespace.

- `description` The description of the container.

- `namespace_id` - The container namespace ID of the container.

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

- `organization_id` - The organization ID the container is associated with.

- `status` - The container status.

- `cron_status` - The cron status of the container.

- `error_message` - The error message of the container.