---
subcategory: "Mailbox"
page_title: "Scaleway: scaleway_mailbox_forwarding"
---

# Resource: scaleway_mailbox_forwarding

Creates and manages a [Scaleway Mailbox](https://www.scaleway.com/en/docs/mailbox/concepts/) [forwarding rule](https://www.scaleway.com/en/developers/api/mailbox/forwardings).

A **forwarding rule** redirects all incoming mail for a `scaleway_mailbox` to a destination
email address. A mailbox can have up to five forwarding rules. The destination must differ
from the mailbox's own address.

Use `forwarding_enabled` / `forwarding_keep_copy` on `scaleway_mailbox` to control whether
forwarding is active and whether a local copy is kept.

~> **Note** This resource is in **alpha** and the API is subject to change.

## Example Usage

### Basic forwarding rule

```terraform
resource "scaleway_mailbox_domain" "main" {
  name = "mail.example.com"
}

resource "scaleway_mailbox" "main" {
  domain_id            = scaleway_mailbox_domain.main.id
  local_part           = "contact"
  password_wo          = var.mailbox_password
  password_wo_version  = 1
  subscription_period  = "monthly"
  forwarding_enabled   = true
  forwarding_keep_copy = true
}

resource "scaleway_mailbox_forwarding" "backup" {
  mailbox_id = scaleway_mailbox.main.id
  email      = "backup@example.com"
}
```

## Argument Reference

- `mailbox_id` - (Required, Forces new resource) UUID of the `scaleway_mailbox` for which to create the rule.
- `email` - (Required) Destination email address. Can be updated in place. Must not be the mailbox's own address.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - Unique identifier of the forwarding rule (UUID).
- `status` - Current status of the rule:
    - `provisioning` – rule is being provisioned.
    - `ready` – rule is active.
    - `updating` – rule is being updated.
    - `deleting` – rule is being deleted.
    - `unknown_status` – status is not known.
- `created_at` - Date and time of rule creation (RFC 3339 format).
- `updated_at` - Date and time of last update (RFC 3339 format).

## Import

Forwarding rules can be imported using their UUID:

```bash
terraform import scaleway_mailbox_forwarding.backup 11111111-1111-1111-1111-111111111111
```
