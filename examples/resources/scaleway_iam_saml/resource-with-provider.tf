### Enable and configure IAM SAML for an organization

resource "scaleway_iam_saml" "with_provider" {
  organization_id    = "11111111-1111-1111-1111-111111111111"
  entity_id          = "https://example.com/saml/metadata"
  single_sign_on_url = "https://example.com/saml/sso"
}
