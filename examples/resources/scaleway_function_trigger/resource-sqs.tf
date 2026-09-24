### SQS

resource "scaleway_function_trigger" "main" {
  function_id = scaleway_function.main.id
  name        = "my-trigger"
  sqs {
    project_id = scaleway_mnq_sqs.main.project_id
    queue      = "MyQueue"
    # If region is different
    region = scaleway_mnq_sqs.main.region
  }
}
