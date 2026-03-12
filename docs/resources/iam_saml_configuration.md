---
subcategory: "IAM"
page_title: "Scaleway: scaleway_iam_saml_configuration"
---

# Resource: scaleway_iam_saml_configuration

Manages SAML configuration parameters for an organization. SAML (Security Assertion Markup Language) is an open standard for exchanging authentication and authorization data between parties, specifically between an identity provider and a service provider. This resource allows you to configure the entity_id and single_sign_on_url for SAML-based single sign-on for your Scaleway organization.

Note: This resource requires that SAML is already enabled using the [`scaleway_iam_saml`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/iam_saml) resource.



## Example Usage

```terraform
### Enable and configure IAM SAML for an organization

# Enable SAML on your Scaleway organization
resource "scaleway_iam_saml" "sp" {
  organization_id = "11111111-1111-1111-1111-111111111111"
}

# Configure your external IdP and retrieve its configuration
# Example output values that you would use in your IdP configuration:
# - Scaleway SAML Entity ID: ${scaleway_iam_saml.sp.service_provider.entity_id}
# - Scaleway SAML ACS URL: ${scaleway_iam_saml.sp.service_provider.assertion_consumer_service_url}

# Configure your organization SAML with your IdP 
resource "scaleway_iam_saml_configuration" "main" {
  organization_id    = scaleway_iam_saml.sp.organization_id
  entity_id          = my_idp.main.entity_id
  single_sign_on_url = my_idp.main.sso_url
}
```



## Argument Reference

The following arguments are supported:

- `organization_id` - (Optional) The organization ID. If not provided, the default organization ID will be used.
- `entity_id` - (Optional) The entity ID of the SAML Identity Provider.
- `single_sign_on_url` - (Optional) The single sign-on URL of the SAML Identity Provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - (Computed) The ID of the SAML configuration.
- `status` - (Computed) The status of the SAML configuration.
