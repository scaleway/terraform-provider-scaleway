# Apply an ephemeral policy on your secret named `foo`.
# In the example below, your secret's lifetime is of 24 hours, your secret versions will expire once they are accessed, and they are disabled after being accessed.

resource "scaleway_secret" "ephemeral" {
  name = "foo"
  ephemeral_policy {
    ttl                   = "24h"
    expires_once_accessed = true
    action                = "disable"
  }
}
