## Retrieve a Serverless Functions namespace

// Get info by namespace name
data "scaleway_function_namespace" "my_namespace" {
  name = "my-namespace-name"
}

// Get info by namespace ID
data "scaleway_function_namespace" "my_namespace" {
  namespace_id = "11111111-1111-1111-1111-111111111111"
}
