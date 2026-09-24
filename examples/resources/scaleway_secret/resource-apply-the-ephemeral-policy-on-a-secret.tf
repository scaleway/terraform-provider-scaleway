### Apply the ephemeral policy on a secret

resource "scaleway_secret" "ephemeral" {
  name = "foo"
  ephemeral_policy {
    ttl                   = "24h"
    expires_once_accessed = true
    action                = "disable"
  }
}
