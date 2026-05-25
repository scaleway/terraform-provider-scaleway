---
subcategory: "Containers"
page_title: "Scaleway: scaleway_container"
---
# scaleway_container

The `scaleway_container` data source is used to retrieve information about a Serverless Container.

Refer to the Serverless Containers [product documentation](https://www.scaleway.com/en/docs/serverless/containers/) and [API documentation](https://www.scaleway.com/en/developers/api/serverless-containers/) for more information.

For more information on the limitations of Serverless Containers, refer to the [dedicated documentation](https://www.scaleway.com/en/docs/serverless-containers/reference-content/containers-limitations/).

## Retrieve a Serverless Container

The following commands allow you to:

- retrieve a container by its name
- retrieve a container by its ID

```hcl
resource "scaleway_container_namespace" "main" {
}

resource "scaleway_container" "main" {
  name         = "test-container-data"
  namespace_id = scaleway_container_namespace.main.id
}

// Get info by container name
data "scaleway_container" "by_name" {
  namespace_id = scaleway_container_namespace.main.id
  name         = scaleway_container.main.name
}

// Get info by container ID
data "scaleway_container" "by_id" {
  namespace_id = scaleway_container_namespace.main.id
  container_id = scaleway_container.main.id
}
```

## Arguments reference

This section lists the arguments that you can provide to the `scaleway_container` data source to filter and retrieve the desired namespace. Each argument has a specific purpose:

- `name` - (Required) The unique name of the container.

- `namespace_id` - (Required) The container namespace ID of the container.

- `project_id` - (Optional) The unique identifier of the project with which the container is associated.

~> **Important** Updating the `name` argument will recreate the container.

## Attributes Reference

The `scaleway_container` data source exports certain attributes once the container information is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the container

~> **Important:** Containers' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `description` The description of the container.

- `tags` - The list of tags associated with the container.

- `environment_variables` - The [environment variables](https://www.scaleway.com/en/docs/serverless-containers/concepts/#environment-variables) of the container.

- `secret_environment_variables` - The [secret environment variables](https://www.scaleway.com/en/docs/serverless-containers/concepts/#secrets) of the container.

- `min_scale` - The minimum number of container instances running continuously.

- `max_scale` - The maximum number of instances this container can scale to.

- `memory_limit_bytes` - The memory resources in bytes to allocate to each container.

- `memory_limit` - (Deprecated) The memory resources in MB to allocate to each container.

- `cpu_limit` - The amount of vCPU computing resources to allocate to each container.

- `timeout` - The maximum amount of time in seconds your container can spend processing a request before being stopped. Default to `300` seconds.

- `privacy` - The privacy type defines the way to authenticate to your container. Please check our dedicated [section](https://www.scaleway.com/en/developers/api/serverless-containers/#protocol-9dd4c8).

- `image` - The image address (e.g., `rg.fr-par.scw.cloud/$NAMESPACE/$IMAGE`)

- `registry_image` - (Deprecated) The registry image address (e.g., `rg.fr-par.scw.cloud/$NAMESPACE/$IMAGE`)

- `protocol` - The communication [protocol](https://www.scaleway.com/en/developers/api/serverless-containers/#path-containers-update-an-existing-container) `http1` or `h2c`. Defaults to `http1`.

- `https_connections_only` - Allows both HTTP and HTTPS (`false`) or redirect HTTP to HTTPS (`true`). Defaults to `false`.

- `http_option` - (Deprecated) Allows both HTTP and HTTPS (`enabled`) or redirect HTTP to HTTPS (`redirected`). Defaults to `enabled`.

- `sandbox` - Execution environment of the container.

- `liveness_probe` - Defines how to check if the container is running.
    - `tcp` - When set to `true`, performs TCP checks on the container.
    - `http` - Perform HTTP check on the container with the specified path.
        - `path` - Path to use for the HTTP health check.
    - `failure_threshold` - Number of consecutive failures before considering the container has to be restarted.
    - `interval`- Time interval between checks (in duration notation, e.g. "30s").
    - `duration` - Duration before the check times out (in duration notation, e.g. "30s").

- `health_check` - (Deprecated) Health check configuration block of the container.
    - `tcp` - When set to `true`, performs TCP checks on the container.
    - `http` - HTTP health check configuration.
        - `path` - Path to use for the HTTP health check.
    - `failure_threshold` - Number of consecutive health check failures before considering the container unhealthy.
    - `interval`- Period between health checks (in seconds).

- ` startup_probe` - Defines how to check if the container has started successfully.
    - `tcp` - When set to `true`, performs TCP checks on the container.
    - `http` - Perform HTTP check on the container with the specified path.
        - `path` - Path to use for the HTTP health check.
    - `failure_threshold` - Number of consecutive failures before considering the container has to be restarted.
    - `interval`- Time interval between checks (in duration notation, e.g. "30s").
    - `duration` - Duration before the check times out (in duration notation, e.g. "30s").

- `scaling_option` - Configuration block used to decide when to scale up or down. Possible values:
    - `concurrent_requests_threshold` - Scale depending on the number of concurrent requests being processed per container instance.
    - `cpu_usage_threshold` - Scale depending on the CPU usage of a container instance.
    - `memory_usage_threshold`- Scale depending on the memory usage of a container instance.

- `port` - The port to expose the container.

- `local_storage_limit_bytes` - Local storage limit of the container (in bytes).

- `local_storage_limit` - (Deprecated) Local storage limit of the container (in MB)

- `command` - Command executed when the container starts. This overrides the default command defined in the container image. This is usually the main executable, or entry point script to run.

- `args` - Arguments passed to the command specified in the "command" field. These override the default arguments from the container image, and behave like command-line parameters.

- `private_network_id` The ID of the Private Network the container is connected to.

- `status` - The container status.

- `cron_status` - The cron status of the container.

- `error_message` - The error message of the container.

- `domain_name` - The native domain name of the container

- `public_endpoint` - The native domain name of the container

- `region` - (Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions) in which the container was created.
