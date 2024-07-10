---
subcategory: "Containers"
page_title: "Scaleway: scaleway_container_token"
---

# Resource: scaleway_container_token

The `scaleway_container_token` resource allows you to create and manage authentication tokens for Scaleway [Serverless Containers](https://www.scaleway.com/en/docs/serverless/containers/).

Refer to the Containers tokens [documentation](https://www.scaleway.com/en/docs/serverless/containers/how-to/create-auth-token-from-console/) and [API documentation](https://www.scaleway.com/en/developers/api/serverless-containers/#path-tokens-list-all-tokens) for more information.

## Add an authentication token to a container

```terraform
resource scaleway_container_namespace main {
  name = "test-container-token-ns"
}

resource scaleway_container main {
  namespace_id = scaleway_container_namespace.main.id
}

// Namespace Token
resource scaleway_container_token namespace {
  namespace_id = scaleway_container_namespace.main.id
  expires_at = "2022-10-18T11:35:15+02:00"
}

// Container Token
resource scaleway_container_token container {
  container_id = scaleway_container.main.id
}
```

## Argument Reference

The following arguments are supported:

- `namespace_id` - (Required) The unique identifier of the Containers namespace.

- `container_id` - (Required) The unique identifier of the container.

~> Only one of `namespace_id` or `container_id` must be set.

- `description` (Optional) The description of the token.

- `expires_at` (Optional) The expiration date of the token.

- `region` - (Defaults to [provider](../index.md#region) `region`). The [region](../guides/regions_and_zones.md#regions) in which the namespace is created.

~> **Important** Updating any of the arguments above will recreate the token.

## Attributes Reference

The `scaleway_container_token` resource exports certain attributes once the authentication token is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

- `id` - The unique identifier of the token.

~> **Important:** Container token IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `token` - The token.

## Import

Tokens can be imported using `{region}/{id}`, as shown below:

```bash
terraform import scaleway_container_token.main fr-par/11111111-1111-1111-1111-111111111111
```
