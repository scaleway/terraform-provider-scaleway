resource "scaleway_function_namespace" "main" {
  name = "test-function-token-ns"
}

resource "scaleway_function" "main" {
  namespace_id = scaleway_function_namespace.main.id
  runtime      = "go118"
  handler      = "Handle"
  privacy      = "private"
}

// Namespace Token
resource "scaleway_function_token" "namespace" {
  namespace_id = scaleway_function_namespace.main.id
  expires_at   = "2022-10-18T11:35:15+02:00"
}

// Function Token
resource "scaleway_function_token" "function" {
  function_id = scaleway_function.main.id
}
