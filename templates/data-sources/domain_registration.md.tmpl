---
subcategory: "Domains and DNS"
page_title: "Scaleway: scaleway_domain_registration"
---

# scaleway_domain_registration

The `scaleway_domain_registration` data source retrieves information about an existing domain registration by domain name. Use it to get the `task_id` and `project_id` needed for importing a domain registration into Terraform state.

Refer to the Domains and DNS [product documentation](https://www.scaleway.com/en/docs/network/domains-and-dns/) and [API documentation](https://www.scaleway.com/en/developers/api/domains-and-dns/) for more information.

## Example Usage

### Get task_id for import

```hcl
data "scaleway_domain_registration" "example" {
  domain_name = "example.com"
}

output "import_command" {
  value = "terraform import scaleway_domain_registration.example ${data.scaleway_domain_registration.example.project_id}/${data.scaleway_domain_registration.example.task_id}"
}
```

### With project_id filter

```hcl
data "scaleway_domain_registration" "example" {
  domain_name = "example.com"
  project_id  = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
}
```

## Argument Reference

- `domain_name` - (Required) The domain name to look up (e.g. example.com).

- `project_id` - (Optional) The project ID to filter by. Defaults to the project ID in the provider configuration.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The resource ID in `project_id/task_id` format.

- `task_id` - The task ID of the domain registration.

- `project_id` - The project ID of the domain registration.

- `domain_names` - List of domain names in the registration.

- `owner_contact_id` - ID of the owner contact.

- `owner_contact` - Owner contact details (when lastname is API-compatible).

- `administrative_contact` - Administrative contact details.

- `technical_contact` - Technical contact details.

- `auto_renew` - Whether auto-renewal is enabled.

- `dnssec` - Whether DNSSEC is enabled.

- `ds_record` - DNSSEC DS record configuration.
