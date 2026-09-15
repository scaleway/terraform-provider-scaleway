resource "scaleway_function_namespace" "main" {
  name        = "main-function-namespace"
  description = "Main function namespace"
  tags        = ["tag1", "tag2"]
}
