# Extract the region from a resource ID
resource "scaleway_secret" "main" {
  name = "my-secret"
}

output "secret_region" {
  value = provider::scaleway::region_from_id(scaleway_secret.main.id, null)
}

# Extract region with validation disabled (for non-Scaleway regions)
output "custom_region" {
  value = provider::scaleway::region_from_id("xx-yyy-1/12345678-1234-1234-1234-123456789012", true)
}
