---
subcategory: "IAM"
page_title: "Scaleway: scaleway_iam_saml"
---

# Resource: scaleway_iam_saml
Manages SAML activation for an organization. SAML (Security Assertion Markup Language) is an open standard for exchanging authentication and authorization data between parties, specifically between an identity provider and a service provider. This resource allows you to enable and disable SAML-based single sign-on for your Scaleway organization.

For configuring SAML parameters (`entity_id` and `single_sign_on_url`), use the [`scaleway_iam_saml_configuration`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/iam_saml_configuration) resource.



## Example Usage

```terraform
### Enable IAM SAML for an organization

resource "scaleway_iam_saml" "main" {
  organization_id = "11111111-1111-1111-1111-111111111111"
}
```



## Argument Reference

The following arguments are supported:

- `organization_id` - (Optional) The organization ID. If not provided, the default organization ID will be used.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the SAML configuration.
- `entity_id` - (Computed) The entity ID of the SAML Identity Provider.
- `single_sign_on_url` - (Computed) The single sign-on URL of the SAML Identity Provider.
- `status` - (Computed) The status of the SAML configuration.
- `service_provider` - (Computed) The Service Provider information. It contains:
    - `entity_id` - (Computed) The entity ID of the Service Provider.
    - `assertion_consumer_service_url` - (Computed) The assertion consumer service URL of the Service Provider.
