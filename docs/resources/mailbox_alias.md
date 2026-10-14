---
subcategory: "Mailbox"
page_title: "Scaleway: scaleway_mailbox_alias"
---

# Resource: scaleway_mailbox_alias

Creates and manages a [Scaleway Mailbox](https://www.scaleway.com/en/docs/mailbox/concepts/) [alias](https://www.scaleway.com/en/developers/api/mailbox/aliases).

An **alias** is an alternative address (`local_part@domain`) that delivers into an existing
`scaleway_mailbox`. It is not a separate mailbox account.

~> **Note** This resource is in **alpha** and the API is subject to change.

## Example Usage

### Basic alias

```terraform
resource "scaleway_mailbox_domain" "main" {
  name = "mail.example.com"
}

resource "scaleway_mailbox" "john" {
  domain_id           = scaleway_mailbox_domain.main.id
  local_part          = "john.doe"
  password_wo         = var.mailbox_password
  password_wo_version = 1
  subscription_period = "monthly"
}

resource "scaleway_mailbox_alias" "contact" {
  mailbox_id  = scaleway_mailbox.john.id
  local_part  = "contact"
  description = "Public contact address"
}
```

## Argument Reference

- `mailbox_id` - (Required, Forces new resource) UUID of the `scaleway_mailbox` to which this alias belongs.
- `local_part` - (Required, Forces new resource) Local part of the alias email address (part before the `@`). Must be 1–64 characters. Changing this value forces a new alias to be created.
- `description` - (Optional) Description of the alias. Can be updated in place. Omit the attribute to leave it empty (do not set an empty string).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - Unique identifier of the alias (UUID).
- `email` - Full email address of the alias (`local_part@domain`).
- `status` - Current status of the alias:
    - `provisioning` – alias is being provisioned.
    - `ready` – alias is active and ready to use.
    - `deleting` – alias is being deleted.
    - `unknown_status` – status is not known.
- `created_at` - Date and time of alias creation (RFC 3339 format).
- `updated_at` - Date and time of last update (RFC 3339 format).

## Import

Aliases can be imported using their UUID:

```bash
terraform import scaleway_mailbox_alias.contact 11111111-1111-1111-1111-111111111111
```
