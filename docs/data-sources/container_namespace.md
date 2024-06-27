---
subcategory: "Containers"
page_title: "Scaleway: scaleway_container_namespace"
---

# scaleway_container_namespace

The `scaleway_container_namespace` data source is used to retrieve information about a Serverless Containers namespace.

Refer to the Serverless Containers [product documentation](https://www.scaleway.com/en/docs/serverless/containers/) and [API documentation](https://www.scaleway.com/en/developers/api/serverless-containers/) for more information.

## Retrieve a Serverless Containers namespace

The following commands allow you to:

- retrieve a namespace by its name
- retrieve a namespace by its ID

```hcl
// Get info by namespace name
data "scaleway_container_namespace" "by_name" {
  name = "my-namespace-name"
}

// Get info by namespace ID
data "scaleway_container_namespace" "by_id" {
  namespace_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

This section lists the arguments that you can provide to the `scaleway_container_namespace` data source to filter and retrieve the desired namespace. Each argument has a specific purpose:

- `name` - (Optional) The name of the namespace. Only one of `name` and `namespace_id` should be specified.

- `namespace_id` - (Optional) The unique identifier of the namespace. Only one of `name` and `namespace_id` should be specified.

- `region` - (Defaults to the region specified in the [provider configuration](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the namespace exists.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The unique identifier of the project with which the namespace is associated.

## Attributes reference

The `scaleway_container_namespace` data source exports certain attributes once the namespace information is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

In addition to the arguments above, the following attributes are exported:

- `id` - The unique identifier of the container namespace.

~> **Important:** Serverless Containers namespace IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are expressed in the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `organization_id` - The unique identifier of the organization with which the namespace is associated.
- `description` - The description of the namespace.
- `environment_variables` - The environment variables of the namespace.
- `registry_endpoint` - The registry endpoint of the namespace.
- `registry_namespace_id` - The unique identifier of the registry namespace of the Serverless Containers namespace.
