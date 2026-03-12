---
subcategory: "IAM"
page_title: "Scaleway: scaleway_iam_saml"
---

# Resource: scaleway_iam_saml

Manages SAML configuration for an organization. This resource allows you to enable, disable and update SAML-based single sign-on for your Scaleway organization.

SAML (Security Assertion Markup Language) is an open standard for exchanging authentication and authorization data between parties, specifically between an identity provider and a service provider. This resource enables you to configure SAML-based single sign-on (SSO) for your Scaleway organization.



## Example Usage

```terraform
### Enable IAM SAML for an organization

resource "scaleway_iam_saml" "main" {
  organization_id = "11111111-1111-1111-1111-111111111111"
}
```

```terraform
### Enable and configure IAM SAML for an organization

resource "scaleway_iam_saml" "with_provider" {
  organization_id    = "11111111-1111-1111-1111-111111111111"
  entity_id          = "https://example.com/saml/metadata"
  single_sign_on_url = "https://example.com/saml/sso"
}
```



## Argument Reference

The following arguments are supported:

- `organization_id` - (Optional) The organization ID.
- `entity_id` - (Optional) The entity ID of the SAML Identity Provider.
- `single_sign_on_url` - (Optional) The single sign-on URL of the SAML Identity Provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the SAML configuration.
- `service_provider` - The Service Provider information.
- `status` - The status of the SAML configuration.
