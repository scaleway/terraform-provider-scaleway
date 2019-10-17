---
layout: "scaleway"
page_title: "Scaleway: scaleway_account_ssh_key"
description: |-
  Manages Scaleway user SSH keys.
---

# scaleway_account_ssh_key

Manages user SSH keys to access servers provisioned on Scaleway.

## Example Usage

```hcl
resource "scaleway_account_ssh_key" "main" {
    name 	   = "main"
    public_key = "<YOUR-PUBLIC-SSH-KEY>"
}
```

## Arguments Reference

The following arguments are supported:

- `name` - (Required) The name of the SSH key.
- `public_key` - (Required) The public SSH key to be added.
- `organization_id` - (Defaults to [provider](../index.html#organization_id) `organization_id`) The ID of the organization the IP is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the SSH key.

## Import

SSH keys can be imported using the `id`, e.g.

```bash
$ terraform import scaleway_account_ssh_key.main 11111111-1111-1111-1111-111111111111
```
