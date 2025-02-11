---
subcategory: "Functions"
page_title: "Scaleway: scaleway_function_token"
---

# Resource: scaleway_function_token

The `scaleway_function_token` resource allows you to create and manage authentication tokens for Scaleway [Serverless Functions](https://www.scaleway.com/en/docs/serverless/functions/).

Refer to the Functions tokens [documentation](https://www.scaleway.com/en/docs/serverless/functions/how-to/create-auth-token-from-console/) and [API documentation](https://www.scaleway.com/en/developers/api/serverless-functions/#path-tokens-list-all-tokens) for more information.

## Example Usage

```terraform
resource scaleway_function_namespace main {
  name = "test-function-token-ns"
}

resource scaleway_function main {
  namespace_id = scaleway_function_namespace.main.id
  runtime      = "go118"
  handler      = "Handle"
  privacy      = "private"
}

// Namespace Token
resource scaleway_function_token namespace {
  namespace_id = scaleway_function_namespace.main.id
  expires_at = "2022-10-18T11:35:15+02:00"
}

// Function Token
resource scaleway_function_token function {
  function_id = scaleway_function.main.id
}
```

## Argument Reference

The following arguments are supported:

- `namespace_id` - (Required) The unique identifier of the Functions namespace.

- `function_id` - (Required) The unique identifier of the function.

~> Only one of `namespace_id` or `function_id` must be set.

- `description` (Optional) The description of the token.

- `expires_at` (Optional) The expiration date of the token.

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which the namespace is created.

~> **Important** Updating any of the arguments above will recreate the token.

## Attributes Reference

The `scaleway_function_token` resource exports certain attributes once the authentication token is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

- `id` - The unique identifier of the token.

~> **Important:** Function token IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `token` - The token.

## Import

Tokens can be imported using `{region}/{id}`, as shown below:

```bash
terraform import scaleway_function_token.main fr-par/11111111-1111-1111-1111-111111111111
```
