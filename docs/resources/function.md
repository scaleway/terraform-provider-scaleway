---
subcategory: "Functions"
page_title: "Scaleway: scaleway_function"
---

# Resource: scaleway_function

The `scaleway_function` resource allows you to create and manage [Serverless Functions](https://www.scaleway.com/en/docs/serverless/functions/).

Refer to the Serverless Functions [product documentation](https://www.scaleway.com/en/docs/serverless/functions/) and [API documentation](https://www.scaleway.com/en/developers/api/serverless-functions/) for more information.

For more information on the limitations of Serverless Functions, refer to the [dedicated documentation](https://www.scaleway.com/en/docs/compute/functions/reference-content/functions-limitations/).

## Example Usage

### Basic

```terraform
resource "scaleway_function_namespace" "main" {
  name        = "main-function-namespace"
  description = "Main function namespace"
}

resource scaleway_function main {
  namespace_id = scaleway_function_namespace.main.id
  runtime      = "go118"
  handler      = "Handle"
  privacy      = "private"
}
```

### With sources and deploy

You can easily create a zip file containing your function (ex: `zip function.zip -r go.mod go.sum handler.go`) to deploy it with Terraform seamlessly. Refer to our [dedicated documentation](https://www.scaleway.com/en/docs/serverless/functions/how-to/package-function-dependencies-in-zip/) for more information on how to package a function into a zip file.

```terraform
resource "scaleway_function_namespace" "main" {
  name        = "main-function-namespace"
  description = "Main function namespace"
}

resource scaleway_function main {
  namespace_id = scaleway_function_namespace.main.id
  runtime      = "go118"
  handler      = "Handle"
  privacy      = "private"
  timeout      = 10
  zip_file     = "function.zip"
  zip_hash     = filesha256("function.zip")
  deploy       = true
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The unique name of the function name.

- `namespace_id` - (Required) The Functions namespace ID of the function.

~> **Important** Updating the `name` argument will recreate the function.

- `description` (Optional) The description of the function.

- `environment_variables` - (Optional) The [environment variables](https://www.scaleway.com/en/docs/compute/functions/concepts/#environment-variables) of the function.

- `secret_environment_variables` - (Optional) The [secret environment variables](https://www.scaleway.com/en/docs/compute/functions/concepts/#secrets) of the function.

- `privacy` - (Optional) The privacy type defines the way to authenticate to your function. Please check our dedicated [section](https://www.scaleway.com/en/developers/api/serverless-functions/#protocol-9dd4c8).

- `runtime` - Runtime of the function. Runtimes can be fetched using [specific route](https://www.scaleway.com/en/developers/api/serverless-functions/#path-functions-get-a-function)

- `min_scale` - (Optional) The minimum number of function instances running continuously. Defaults to 0. Functions are billed when executed, and using a `min_scale` greater than 0 will cause your function to run constantly.

- `max_scale` - (Optional) The maximum number of instances this function can scale to. Default to 20. Your function will scale automatically based on the incoming workload, but will never exceed the configured `max_scale` value.

- `memory_limit` - (Optional) The memory resources in MB to allocate to each function. Defaults to 128 MB.

- `handler` - Handler of the function, depends on the runtime. Refer to the [dedicated documentation](https://www.scaleway.com/en/developers/api/serverless-functions/#path-functions-create-a-new-function) for the list of supported runtimes.

- `timeout` - (Optional) The maximum amount of time your function can spend processing a request before being stopped. Defaults to 300s.

- `zip_file` - Path to the zip file containing your function sources to upload.

- `zip_hash` - The hash of your source zip file, changing it will redeploy the function. Can be any string, changing it will simply trigger a state change. You can use any Terraform hash function to trigger a change on your zip change (see examples).

- `deploy` - Define whether the function should be deployed. Terraform will wait for the function to be deployed. Your function will be redeployed if you update the source zip file.

<<<<<<< HEAD
- `sandbox` - (Optional) Execution environment of the function.

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which the namespace should be created.
=======
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the function was created.
>>>>>>> adba6efa (docs(review): update)

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the functions namespace is associated with.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier of the function.

~> **Important:** Function IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`.

- `namespace_id` - The namespace ID the function is associated with.

- `domain_name` - The native domain name of the function.

- `organization_id` - The organization ID the function is associated with.

- `cpu_limit` - The CPU limit in mVCPU for your function.

## Import

Functions can be imported using, `{region}/{id}`, as shown below:

```bash
<<<<<<< HEAD
terraform import scaleway_function.main fr-par/11111111-1111-1111-1111-111111111111
```
=======
$ terraform import scaleway_function.main fr-par/11111111-1111-1111-1111-111111111111
```
>>>>>>> adba6efa (docs(review): update)
