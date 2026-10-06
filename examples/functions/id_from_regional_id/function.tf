# Extract the ID without region from a regional ID
resource "scaleway_secret" "main" {
  name = "my-secret"
}

output "secret_id" {
  value = provider::scaleway::id_from_regional_id(scaleway_secret.main.id, null)
}

# Extract ID with validation disabled (for non-Scaleway regions)
output "custom_id" {
  value = provider::scaleway::id_from_regional_id("xx-yyy-1/12345678-1234-1234-1234-123456789012", true)
}
