---
subcategory: "Domains and DNS"
page_title: "Scaleway: scaleway_domain_external_domain_validated"
---

# Resource: scaleway_domain_external_domain_validated

The `scaleway_domain_external_domain_validated` resource allows you to wait until an external domain is validated by Scaleway.
This resource is used to block the Terraform execution until the validation token has been correctly added to the external domain's DNS records and verified by Scaleway.


~> **WARNING:** This resource implements a part of the validation workflow. It does not represent a real-world entity in Scaleway, therefore changing or deleting this resource on its own has no effect. It must be used with a `scaleway_domain_external_domain` resource. 


Once the domain is validated, its status becomes `active`, and it can be used within Scaleway services.

Refer to the [Domains and DNS documentation](https://www.scaleway.com/en/docs/network/domains-and-dns/) and the [API documentation](https://developers.scaleway.com/) for more details.

## Example Usage

### Wait for an External Domain to be Validated

The following example registers an external domain and then waits for it to be validated.

```terraform

resource "scaleway_domain_external_domain" "example" {
  domain = "scw.example.com"
}
resource "scaleway_domain_record" "validation" {
  dns_zone= "example.com"
  name = "_scaleway-challenge.scw"
  type = "TXT"
  data = scaleway_domain_external_domain.example.validation_token
}

# Wait for domain validation (default timeout is 30 minutes)
resource "scaleway_domain_external_domain_validated" "example" {
  domain = scaleway_domain_external_domain.example.domain
}

# After validation, we can use the domain's NS servers
resource "scaleway_domain_record" "NS" {
  for_each = {ns0=scaleway_domain_external_domain_validated.example.ns_servers[0],
    ns1=scaleway_domain_external_domain_validated.example.ns_servers[1]}
  dns_zone= "example.com"
  name = "scw"
  type = "NS"
  data = each.value
}

```

## Argument Reference

The following arguments are supported:

- `domain` (Required, String): The domain name to be validated.
- `project_id` (Optional, String): The Scaleway project ID.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id`: The domain name.
- `ns_servers`: List of NS servers for the domain once it is validated.

## Timeouts

The `timeouts` block allows you to specify a custom duration for waiting until the domain is validated.
Default timeout is 30 minutes.

- `create`: Custom duration for the validation process.


```bash
terraform import scaleway_domain_external_domain_validated.example example.com
```
