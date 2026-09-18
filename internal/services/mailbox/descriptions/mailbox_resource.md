Creates and manages a Scaleway Mailbox mailbox.

A mailbox is a hosted email address (`local_part@domain`) on a `scaleway_mailbox_domain`.
Prefer `password_wo` so the password is never stored in Terraform state.
