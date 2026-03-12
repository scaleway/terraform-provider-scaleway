# Project to be referenced in the IAM policy
data "scaleway_account_project" "default" {
  name = "default"
}

# IAM resources
resource "scaleway_iam_application" "container_auth" {
  name = "container-auth"
}
resource "scaleway_iam_policy" "access_private_containers" {
  application_id = scaleway_iam_application.container_auth.id
  rule {
    project_ids          = [data.scaleway_account_project.default.id]
    permission_set_names = ["ContainersPrivateAccess"]
  }
}
resource "scaleway_iam_api_key" "api_key" {
  application_id = scaleway_iam_application.container_auth.id
}

# Container resources
resource "scaleway_container_namespace" "private" {
  name = "private-container-namespace"
}
resource "scaleway_container" "private" {
  namespace_id   = scaleway_container_namespace.private.id
  registry_image = "rg.fr-par.scw.cloud/my-registry-ns/my-image:latest"
  privacy        = "private"
  deploy         = true
}

# Output the secret key and the container's endpoint for the curl command
output "secret_key" {
  value     = scaleway_iam_api_key.api_key.secret_key
  sensitive = true
}
output "container_endpoint" {
  value = scaleway_container.private.domain_name
}

# Then you can access your private container using the API key:
# $ curl -H "X-Auth-Token: $(terraform output -raw secret_key)" \
#   "https://$(terraform output -raw container_endpoint)/"

# Keep in mind that you should revoke your legacy JWT tokens to ensure maximum security.
