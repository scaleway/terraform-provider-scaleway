---
subcategory: "Containers"
page_title: "Scaleway: scaleway_container_namespace"
---

# Resource: scaleway_container_namespace

The `scaleway_container_namespace` resource allows you to
for Scaleway [Serverless Containers](https://www.scaleway.com/en/docs/serverless/containers/).

Refer to the Containers namespace [documentation](https://www.scaleway.com/en/docs/serverless/containers/how-to/create-a-containers-namespace/) and [API documentation](https://www.scaleway.com/en/developers/api/serverless-containers/#path-namespaces-list-all-your-namespaces) for more information.

## Create a Containers namespace

```terraform
resource "scaleway_container_namespace" "main" {
  name        = "main-container-namespace"
  description = "Main container namespace"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The unique name of the Containers namespace.

~> **Important** Updates to the `name` argument will recreate the namespace.

- `description` (Optional) The description of the namespace.

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which the namespace is created.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The unique identifier of the project that contains the namespace.

- `environment_variables` - The environment variables of the namespace.

- `secret_environment_variables` - The secret environment variables of the namespace.

## Attributes Reference

The `scaleway_container_namespace` resource exports certain attributes once the Containers namespace has been created. These attributes can be referenced in other parts of your Terraform configuration.

- `id` - The unique identifier of the namespace.

~> **Important:** Containers namespace IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`.

- `organization_id` - The Organization ID with which the namespace is associated.
- `registry_endpoint` - The registry endpoint of the namespace.
- `registry_namespace_id` - The registry namespace ID of the namespace.


## Import

Containers namespaces can be imported using `{region}/{id}`, as shown below:

```bash
terraform import scaleway_container_namespace.main fr-par/11111111-1111-1111-1111-111111111111
```
