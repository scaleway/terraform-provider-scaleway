### With sources and deploy

resource "scaleway_function_namespace" "main" {
  name        = "main-function-namespace"
  description = "Main function namespace"
}

resource "scaleway_function" "main" {
  namespace_id = scaleway_function_namespace.main.id
  description  = "function with zip file"
  tags         = ["tag1", "tag2"]
  runtime      = "go124"
  handler      = "Handle"
  privacy      = "private"
  timeout      = 10
  zip_file     = "function.zip"
  zip_hash     = filesha256("function.zip")
  deploy       = true
}
