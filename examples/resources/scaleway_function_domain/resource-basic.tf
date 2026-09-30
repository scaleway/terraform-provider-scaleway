resource "scaleway_function_domain" "main" {
  function_id = scaleway_function.main.id
  hostname    = "example.com"

  depends_on = [
    scaleway_function.main,
  ]
}

resource "scaleway_function_namespace" "main" {}

resource "scaleway_function" "main" {
  namespace_id = scaleway_function_namespace.main.id
  runtime      = "go118"
  privacy      = "private"
  handler      = "Handle"
  zip_file     = "testfixture/gofunction.zip"
  deploy       = true
}
