resource "scaleway_function_namespace" "main" {
  name = "test-cron"
}

resource "scaleway_function" "main" {
  name         = "test-cron"
  namespace_id = scaleway_function_namespace.main.id
  runtime      = "node14"
  privacy      = "private"
  handler      = "handler.handle"
}

resource "scaleway_function_cron" "main" {
  name        = "test-cron"
  function_id = scaleway_function.main.id
  schedule    = "0 0 * * *"
  args        = jsonencode({ test = "scw" })
}

resource "scaleway_function_cron" "func" {
  function_id = scaleway_function.main.id
  schedule    = "0 1 * * *"
  args        = jsonencode({ my_var = "terraform" })
}
