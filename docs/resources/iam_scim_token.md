---
subcategory: "IAM"
page_title: "Scaleway: scaleway_iam_scim_token"
---

# Resource: scaleway_iam_scim_token
SCIM Token resource allows you to create and manage SCIM tokens for SCIM configurations.

SCIM tokens are used to authenticate to SCIM endpoints for automated user provisioning and management. Each token is associated with a specific SCIM configuration and has an expiration time.

> **Note:** The bearer token is only available at creation time and is marked as sensitive. It cannot be retrieved after creation. If you lose the token, you will need to create a new one.



## Example Usage

```terraform
### Basic SCIM Token creation

# Enable SCIM for your organization
resource "scaleway_iam_scim" "main" {
  organization_id = "your-organization-id"
}

# Create a SCIM token
resource "scaleway_iam_scim_token" "main" {
  scim_id         = scaleway_iam_scim.main.id
  organization_id = "your-organization-id"
}
```



## Argument Reference

- `scim_id` - (Required) The SCIM configuration ID for which to create the token.
- `organization_id` - (Optional) The organization ID. If not provided, the default organization configured in the provider is used.

## Attributes Reference

- `id` - The ID of the SCIM token
- `bearer_token` - The Bearer Token to use to authenticate to SCIM endpoints.
- `created_at` - The date and time of SCIM token creation
- `expires_at` - The date and time when the SCIM token expires

## Import

SCIM token can be imported using either:

- Just the token ID (uses default organization):
```bash
terraform import scaleway_iam_scim_token.main 11111111-1111-1111-1111-111111111111
```

- Or with explicit organization ID:
```bash
terraform import scaleway_iam_scim_token.main 22222222-1111-1111-1111-111111111111/11111111-1111-1111-1111-111111111111
```
