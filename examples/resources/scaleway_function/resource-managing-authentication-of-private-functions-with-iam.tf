### Managing authentication of private functions with IAM

# Project to be referenced in the IAM policy
data "scaleway_account_project" "default" {
  name = "default"
}

# IAM resources
resource "scaleway_iam_application" "func_auth" {
  name = "function-auth"
}
resource "scaleway_iam_policy" "access_private_funcs" {
  application_id = scaleway_iam_application.func_auth.id
  rule {
    project_ids          = [data.scaleway_account_project.default.id]
    permission_set_names = ["FunctionsPrivateAccess"]
  }
}
resource "scaleway_iam_api_key" "api_key" {
  application_id = scaleway_iam_application.func_auth.id
}

# Function resources
resource "scaleway_function_namespace" "private" {
  name = "private-function-namespace"
}
resource "scaleway_function" "private" {
  namespace_id = scaleway_function_namespace.private.id
  runtime      = "go124"
  handler      = "Handle"
  privacy      = "private"
  zip_file     = "function.zip"
  zip_hash     = filesha256("function.zip")
  deploy       = true
}

# Output the secret key and the function's endpoint for the curl command
output "secret_key" {
  value     = scaleway_iam_api_key.api_key.secret_key
  sensitive = true
}
output "function_endpoint" {
  value = scaleway_function.private.domain_name
}
