### Basic

resource "scaleway_function_namespace" "main" {
  name        = "main-function-namespace"
  description = "Main function namespace"
}

resource "scaleway_function" "main" {
  namespace_id = scaleway_function_namespace.main.id
  runtime      = "go124"
  handler      = "Handle"
  privacy      = "private"
}
