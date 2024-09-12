---
subcategory: "Functions"
page_title: "Scaleway: scaleway_function_namespace"
---

# scaleway_function_namespace

The `scaleway_function_namespace` data source is used to retrieve information about a Serverless Functions namespace.

Refer to the Serverless Functions [product documentation](https://www.scaleway.com/en/docs/serverless/functions/) and [API documentation](https://www.scaleway.com/en/developers/api/serverless-functions/) for more information.

## Retrieve a Serverless Functions namespace

The following commands allow you to:

- retrieve a namespace by its name
- retrieve a namespace by its ID

```hcl
// Get info by namespace name
data "scaleway_function_namespace" "my_namespace" {
  name = "my-namespace-name"
}

// Get info by namespace ID
data "scaleway_function_namespace" "my_namespace" {
  namespace_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

This section lists the arguments that you can provide to the `scaleway_function_namespace` data source to filter and retrieve the desired namespace. Each argument has a specific purpose:

- `name` - (Optional) The name of the namespace. Only one of `name` and `namespace_id` should be specified.

- `namespace_id` - (Optional) The unique identifier of the namespace. Only one of `name` and `namespace_id` should be specified.

- `region` - (Defaults to the region specified in the [provider configuration](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the namespace exists.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The unique identifier of the project with which the namespace is associated.

## Attributes Reference

The `scaleway_function_namespace` data source exports certain attributes once the namespace information is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

In addition to the arguments above, the following attributes are exported:

- `id` - The unique identifier of the function namespace.

~> **Important:** Serverless Functions namespace IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are expressed in the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `organization_id` - The unique identifier of the organization with which the namespace is associated.
- `description` - The description of the namespace.
- `environment_variables` - The environment variables of the namespace.
- `registry_endpoint` - The registry endpoint of the namespace.
- `registry_namespace_id` - The unique identifier of the registry namespace of the Serverless Functions namespace.
