---
subcategory: "Functions"
page_title: "Scaleway: scaleway_function_domain"
---

# Resource: scaleway_function_domain

The `scaleway_function_domain` resource allows you to create and manage domain name bindings for Scaleway [Serverless Functions](https://www.scaleway.com/en/docs/serverless/functions/).

Refer to the Functions domain [documentation](https://www.scaleway.com/en/docs/serverless/functions/how-to/add-a-custom-domain-name-to-a-function/) and the [API documentation](https://www.scaleway.com/en/developers/api/serverless-functions/#path-domains-list-all-domain-name-bindings) for more information.

## Example Usage

This command allows to bind a custom domain name to a function.

```terraform
resource "scaleway_function_domain" "main" {
  function_id = scaleway_function.main.id
  hostname    = "example.com"

  depends_on = [
    scaleway_function.main,
  ]
}

resource scaleway_function_namespace main {}

resource scaleway_function main {
  namespace_id = scaleway_function_namespace.main.id
  runtime = "go118"
  privacy = "private"
  handler = "Handle"
  zip_file = "testfixture/gofunction.zip"
  deploy = true
}
```

## Argument Reference

The following arguments are supported:

- `function_id` - (Required) The unique identifier of the function.

- `hostname` - (Required) The hostname with a CNAME record.

  We recommend you use a CNAME domain record that point to your native function `domain_name` for it.

~> **Important** Updating the `function_id` or `hostname` arguments will recreate the domain.

## Attributes Reference

The `scaleway_function_domain` resource exports certain attributes once the function domain name has been retrieved. These attributes can be referenced in other parts of your Terraform configuration.

- `id` - The unique identifier of the function domain.

~> **Important:** Function domain IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the domain was created.

- `url` - The URL used to query the function.

## Import

Function domain binding can be imported using `{region}/{id}`, as shown below:

```bash
terraform import scaleway_function_domain.main fr-par/11111111-1111-1111-1111-111111111111
```
