---
subcategory: "IAM"
page_title: "Scaleway: scaleway_iam_scim"
---

# Resource: scaleway_iam_scim
SCIM (System for Cross-domain Identity Management) resource allows you to enable or disable SCIM for an organization.

SCIM is a standard for automating the exchange of user identity information between identity domains, or IT systems. When enabled, it allows for automated provisioning and deprovisioning of user accounts.



## Example Usage

```terraform
### Enable IAM SCIM for an organization

resource "scaleway_iam_scim" "main" {
  organization_id = "11111111-1111-1111-1111-111111111111"
}
```



## Argument Reference

- `organization_id` - (Optional) The organization ID. If not provided, the default organization configured in the provider is used.

## Attributes Reference

- `id` - The ID of the SCIM configuration.
- `created_at` - The date and time of SCIM configuration creation.

## Import

SCIM can be imported using the organization ID:

```bash
terraform import scaleway_iam_scim.main 11111111-1111-1111-1111-111111111111
```
