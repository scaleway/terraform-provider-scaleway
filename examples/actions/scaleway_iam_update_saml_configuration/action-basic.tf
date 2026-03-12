### Enable and configure IAM SAML with an external Identity Provider

# Enable SAML on your Scaleway organization
resource "scaleway_iam_saml" "sp" {
  organization_id = "11111111-1111-1111-1111-111111111111"
}

# Configure your external IdP and retrieve its configuration
# Example output values that you would use in your IdP configuration:
# - Scaleway SAML Entity ID: ${scaleway_iam_saml.sp.service_provider.entity_id}
# - Scaleway SAML ACS URL: ${scaleway_iam_saml.sp.service_provider.assertion_consumer_service_url}

# Configure your organization SAML with your configurations from your IdP
action "scaleway_iam_update_saml_configuration" "main" {
  organization_id    = scaleway_iam_saml.sp.organization_id
  entity_id          = my_idp.main.entity_id
  single_sign_on_url = my_idp.main.sso_url
}
