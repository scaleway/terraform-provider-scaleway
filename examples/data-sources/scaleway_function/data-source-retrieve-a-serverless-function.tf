## Retrieve a Serverless Function

// Get info by function name
data "scaleway_function" "my_function" {
  name         = "my-namespace-name"
  namespace_id = "11111111-1111-1111-1111-111111111111"
}

// Get info by function ID
data "scaleway_function" "my_function" {
  function_id  = "11111111-1111-1111-1111-111111111111"
  namespace_id = "11111111-1111-1111-1111-111111111111"
}
