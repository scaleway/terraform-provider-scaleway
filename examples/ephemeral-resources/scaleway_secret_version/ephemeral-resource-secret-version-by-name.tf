### Create a secret version and access it by its name without ever persisting the data in the state file

resource "scaleway_secret" "main" {
  name        = "my-secret"
  description = "my-secret-description"
}

resource "scaleway_secret_version" "v1" {
  description     = "version1"
  secret_id       = scaleway_secret.main.id
  data_wo         = "my_super_secret_data"
  data_wo_version = 1
  depends_on      = [scaleway_secret.main]
}

# Access a specific secret version revision using the ephemeral resource with secret_name
ephemeral "scaleway_secret_version" "data_v1" {
  secret_name = scaleway_secret.main.name
  revision    = "1"
  depends_on  = [scaleway_secret_version.v1]
}

# Access the latest secret version using the ephemeral resource
ephemeral "scaleway_secret_version" "data_latest" {
  secret_name = scaleway_secret.main.name
  revision    = "latest"
  depends_on  = [scaleway_secret_version.v1]
}
