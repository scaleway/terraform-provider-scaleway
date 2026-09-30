## Retrieve a Serverless Containers namespace

// Get info by namespace name
data "scaleway_container_namespace" "by_name" {
  name = "my-namespace-name"
}

// Get info by namespace ID
data "scaleway_container_namespace" "by_id" {
  namespace_id = "11111111-1111-1111-1111-111111111111"
}
