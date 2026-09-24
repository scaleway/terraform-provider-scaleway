### NATS

resource "scaleway_function_trigger" "main" {
  function_id = scaleway_function.main.id
  name        = "my-trigger"
  nats {
    account_id = scaleway_mnq_nats_account.main.id
    subject    = "MySubject"
    # If region is different
    region = scaleway_mnq_nats_account.main.region
  }
}
