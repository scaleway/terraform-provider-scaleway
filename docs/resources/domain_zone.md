---
subcategory: "Domains and DNS"
page_title: "Scaleway: scaleway_domain_zone"
---

# Resource: scaleway_domain_zone

The `scaleway_domain_zone` resource allows you to create and manage Scaleway DNS zones.

Refer to the Domains and DNS [product documentation](https://www.scaleway.com/en/docs/network/domains-and-dns/) and [API documentation](https://www.scaleway.com/en/developers/api/domains-and-dns/) for more information.

## Example Usage

### Create a DNS zone

The following command allows you to create a DNS zone for the `test.scaleway-terraform.com` subdomain.


```terraform
resource "scaleway_domain_zone" "test" {
  domain    = "scaleway-terraform.com"
  subdomain = "test"
}
```

## Arguments reference

The following arguments are supported:

- `domain` - (Required) The main domain where the DNS zone will be created.

- `subdomain` - (Required) The name of the subdomain (zone name) to create within the domain.

- `project_id` - (Defaults to Project ID specified in the [provider configuration](../index.md#project_id) `project_id`) The ID of the Project associated with the domain.


## Attributes reference

This section lists the attributes that are exported when the `scaleway_domain_zone` resource is created:

- `id` - The ID of the zone, in the `{subdomain}.{domain}` format.

- `ns` - The list of same servers for the zone.

- `ns_default` -  The default list of same servers for the zone.

- `ns_master` - The master list of same servers for the zone.

- `status` - The status of the domain zone.

- `message` - Message.

- `updated_at` - The date and time at which the DNS zone was last updated.

## Import

This section explains how to import a zone using the `{subdomain}.{domain}` format.

```bash
terraform import scaleway_domain_zone.test test.scaleway-terraform.com
```
