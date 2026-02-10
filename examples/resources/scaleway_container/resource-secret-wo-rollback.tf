### Create a container with Write Only secret environment variables (not stored in state), update the secrets, and rollback, using Scaleway Secrets while ensuring the secrets are never stored in the state

resource "scaleway_container_namespace" "main" {
  name        = "my-ns-test"
  description = "test container"
}

# Generate an ephemeral random password (not stored in the state)
ephemeral "random_password" "main" {
  length      = 20
  special     = true
  upper       = true
  lower       = true
  numeric     = true
  min_upper   = 1
  min_lower   = 1
  min_numeric = 1
  min_special = 1
  # Exclude characters that might cause issues in some contexts
  override_special = "!@#$%^&*()_+-=[]{}|;:,.<>?"
}

# Create a secret to store the generated data. We will call it a pretend API key for this example 
resource "scaleway_secret" "api_key" {
  name        = "container-api-key"
  description = "API key for container"
}

# Store the generated API key in a Write Only secret version (not stored in the state)
resource "scaleway_secret_version" "api_key_v1" {
  secret_id       = scaleway_secret.api_key.id
  data_wo         = ephemeral.random_password.main.result
  data_wo_version = 1
}

# Create a container with initial secrets
resource "scaleway_container" "main" {
  name            = "my-container-wo"
  description     = "write-only secret environment variables rollback test"
  tags            = ["tag1", "tag2"]
  namespace_id    = scaleway_container_namespace.main.id
  registry_image  = "${scaleway_container_namespace.main.registry_endpoint}/alpine:test"
  port            = 9997
  cpu_limit       = 1024
  memory_limit    = 2048
  min_scale       = 3
  max_scale       = 5
  timeout         = 600
  max_concurrency = 80
  privacy         = "private"
  protocol        = "http1"
  deploy          = true

  command = ["bash", "-c", "script.sh"]
  args    = ["some", "args"]

  environment_variables = {
    "foo" = "var"
  }
  secret_environment_variables_wo = {
    "API_KEY"     = ephemeral.random_password.main.result
    "DB_PASSWORD" = "initial_password"
  }
  secret_environment_variables_wo_version = 1
}

## Generate a new ephemeral API key for update (not stored in the state)
# ephemeral "random_password" "updated" {
#   length      = 20
#   special     = true
#   upper       = true
#   lower       = true
#   numeric     = true
#   min_upper   = 1
#   min_lower   = 1
#   min_numeric = 1
#   min_special = 1
#   # Exclude characters that might cause issues in some contexts
#   override_special = "!@#$%^&*()_+-=[]{}|;:,.<>?"
# }

## Store the updated API key in a new Write Only secret version (not stored in the state)
# resource "scaleway_secret_version" "api_key_v2" {
#   secret_id       = scaleway_secret.api_key.id
#   data_wo         = ephemeral.random_password.updated.result
#   data_wo_version = 2
# }

## Update the container secrets to new values
# resource "scaleway_container" "main" {
#   name            = "my-container-wo"
#   description     = "write-only secret environment variables rollback test"
#   tags            = ["tag1", "tag2"]
#   namespace_id    = scaleway_container_namespace.main.id
#   registry_image  = "${scaleway_container_namespace.main.registry_endpoint}/alpine:test"
#   port            = 9997
#   cpu_limit       = 1024
#   memory_limit    = 2048
#   min_scale       = 3
#   max_scale       = 5
#   timeout         = 600
#   max_concurrency = 80
#   privacy         = "private"
#   protocol        = "http1"
#   deploy          = true
#
#   command = ["bash", "-c", "script.sh"]
#   args    = ["some", "args"]
#
#   environment_variables = {
#     "foo" = "var"
#   }
#   secret_environment_variables_wo = {
#     "API_KEY" = ephemeral.random_password.updated.result
#     "DB_PASSWORD" = "updated_password"
#   }
#   secret_environment_variables_wo_version = 2
# }

## Query the first API key version as an Ephemeral Resource (not stored in the state)
# ephemeral "scaleway_secret_version" "api_key_v1" {
#   secret_id = scaleway_secret.api_key.id
#   revision   = 1
# }

## Rollback the container API key to the first version
# resource "scaleway_container" "main" {
#   name            = "my-container-wo"
#   description     = "write-only secret environment variables rollback test"
#   tags            = ["tag1", "tag2"]
#   namespace_id    = scaleway_container_namespace.main.id
#   registry_image  = "${scaleway_container_namespace.main.registry_endpoint}/alpine:test"
#   port            = 9997
#   cpu_limit       = 1024
#   memory_limit    = 2048
#   min_scale       = 3
#   max_scale       = 5
#   timeout         = 600
#   max_concurrency = 80
#   privacy         = "private"
#   protocol        = "http1"
#   deploy          = true
#
#   command = ["bash", "-c", "script.sh"]
#   args    = ["some", "args"]
#
#   environment_variables = {
#     "foo" = "var"
#   }
#   secret_environment_variables_wo = {
#     "API_KEY" = ephemeral.scaleway_secret_version.api_key_v1.data
#     "DB_PASSWORD" = "initial_password"
#   }
#   secret_environment_variables_wo_version = 1
# }
