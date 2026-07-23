### Basic

resource "scaleway_mnq_nats_account" "main" {
  name = "nats-account"
}

resource "scaleway_mnq_nats_credentials" "main" {
  account_id = scaleway_mnq_nats_account.main.id
}
