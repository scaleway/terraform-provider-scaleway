---
layout: "scaleway"
page_title: "Scaleway: scaleway_account_ssh_key"
description: |-
  Get information on a Scaleway SSH key.
---

# scaleway_account_ssh_key

Use this data source to get SSH key information based on its ID or name.

## Example Usage

```hcl
// Get info by ssh key name
data "scaleway_account_ssh_key" "my_key" {
  name  = "my-key-name"
}

// Get info by ssh key id
data "scaleway_account_ssh_key" "my_key" {
  ssh_key_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - The ssh key name. Only one of `name` and `ssh_key_id` should be specified.
- `ssh_key_id` - The ssh key id. Only one of `name` and `ssh_key_id` should be specified.
- `organization_id` - (Defaults to [provider](../index.html#organization_id) `organization_id`) The ID of the organization the server is associated with.


## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the server.
- `public_key` - The ssh public key string
