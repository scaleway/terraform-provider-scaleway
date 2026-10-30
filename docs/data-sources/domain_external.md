---
subcategory: "Domains and DNS"
page_title: "Scaleway: scaleway_domain_external"
---

# scaleway_domain_external

The `scaleway_domain_external` data source is used to retrieve information about an external domain registered to be managed by Scaleway.

Refer to the Domains and DNS [product documentation](https://www.scaleway.com/en/docs/network/domains-and-dns/) and [API documentation](https://www.scaleway.com/en/developers/api/domains-and-dns/) for more information.

## Example Usage

### Retrieve an external domain

The following example retrieves the information of the external domain `example.com`.

```terraform
data "scaleway_domain_external" "example" {
  domain = "example.com"
}

output "ns_servers" {
  value = data.scaleway_domain_external.example.ns_servers
}
```

## Argument Reference

The following arguments are supported:

- `domain` - (Required) The domain name to look up.

- `organization_id` - (Optional) The organization ID the external domain belongs to. If not provided, the default organization configured in the provider is used.

- `project_id` - (Defaults to Project ID specified in [the provider configuration](../index.md#arguments-reference)). The ID of the Scaleway Project the external domain belongs to. If not provided, the default project id configured in the provider is used.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The domain name.

- `ns_servers` - List of NS servers for the domain once it is validated.

- `status` - The status of the domain.
