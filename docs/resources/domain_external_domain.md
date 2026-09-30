---
subcategory: "Domains and DNS"
page_title: "Scaleway: scaleway_domain_external_domain"
---

# Resource: scaleway_domain_external_domain

The `scaleway_domain_external_domain` resource allows you to register an external domain name to be managed by Scaleway.
This is useful when you want to use Scaleway DNS for a domain registered with another registrar.

Once registered, you will receive a `validation_token` that you must add as a TXT record to your domain's current DNS configuration to prove ownership.
After ownership is verified, you can point your domain's NS records to the `ns_servers` provided by Scaleway.

Refer to the [Domains and DNS documentation](https://www.scaleway.com/en/docs/network/domains-and-dns/) and the [API documentation](https://developers.scaleway.com/) for more details.

## Example Usage

### Register an External Domain

The following example registers an external domain and retrieves its validation token.

```terraform
resource "scaleway_domain_external_domain" "example" {
  domain = "example.com"
}

output "validation_token" {
  value = scaleway_domain_external_domain.example.validation_token
}
```

## Argument Reference

The following arguments are supported:

- `domain` (Required, String): The domain name to be registered.
- `project_id` (Optional, String): The Scaleway project ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id`: The domain name.
- `ns_servers`: List of NS servers for the domain once it is validated.
- `status`: The status of the domain (e.g., `checking`, `active`).
- `validation_token`: The validation token to be added as a TXT record to the domain's DNS to verify ownership.

## Import

To import an existing external domain, use:

```bash
terraform import scaleway_domain_external_domain.example example.com
```
