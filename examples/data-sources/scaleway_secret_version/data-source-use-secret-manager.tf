### Use Secret Manager

# Create a secret named fooii
resource "scaleway_secret" "main" {
  name        = "fooii"
  description = "barr"
}

# Create a version of fooii containing data
resource "scaleway_secret_version" "main" {
  description = "your description"
  secret_id   = scaleway_secret.main.id
  data        = "your_secret"
}

# Retrieve the secret version specified by the secret ID and the desired version
data "scaleway_secret_version" "data_by_secret_id" {
  secret_id  = scaleway_secret.main.id
  revision   = "1"
  depends_on = [scaleway_secret_version.main]
}

# Retrieve the secret version specified by the secret name and the desired version
data "scaleway_secret_version" "data_by_secret_name" {
  secret_name = scaleway_secret.main.name
  revision    = "1"
  depends_on  = [scaleway_secret_version.main]
}

# Display sensitive data
output "scaleway_secret_access_payload" {
  value = data.scaleway_secret_version.data_by_secret_name.data
}

# Display sensitive data
output "scaleway_secret_access_payload_by_id" {
  value = data.scaleway_secret_version.data_by_secret_id.data
}
