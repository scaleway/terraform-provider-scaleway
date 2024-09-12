---
subcategory: "Functions"
page_title: "Scaleway: scaleway_function_namespace"
---

# Resource: scaleway_function_namespace

The `scaleway_function_namespace` resource allows you to
for Scaleway [Serverless Functions](https://www.scaleway.com/en/docs/serverless/functions/).

Refer to the Functions namespace [documentation](https://www.scaleway.com/en/docs/serverless/functions/how-to/create-a-functions-namespace/) and [API documentation](https://www.scaleway.com/en/developers/api/serverless-functions/#path-namespaces-list-all-your-namespaces) for more information.

## Example Usage

```terraform
resource "scaleway_function_namespace" "main" {
  name        = "main-function-namespace"
  description = "Main function namespace"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The unique name of the Functions namespace.

~> **Important** Updates to the `name` argument will recreate the namespace.

- `description` (Optional) The description of the namespace.

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which the namespace is created.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The unique identifier of the project that contains the namespace.

- `environment_variables` - The environment variables of the namespace.

- `secret_environment_variables` - The secret environment variables of the namespace.

## Attributes Reference

The `scaleway_function_namespace` resource exports certain attributes once the Functions namespace has been created. These attributes can be referenced in other parts of your Terraform configuration.

- `id` - The unique identifier of the namespace.

~> **Important:** Functions namespace IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`.

- `organization_id` - The Organization ID with which the namespace is associated.

- `registry_endpoint` - The registry endpoint of the namespace.

- `registry_namespace_id` - The registry namespace ID of the namespace.

## Import

Functions namespaces can be imported using `{region}/{id}`, as shown below:

```bash
<<<<<<< HEAD
terraform import scaleway_function_namespace.main fr-par/11111111-1111-1111-1111-111111111111
```
=======
$ terraform import scaleway_function_namespace.main fr-par/11111111-1111-1111-1111-111111111111
```
>>>>>>> adba6efa (docs(review): update)
