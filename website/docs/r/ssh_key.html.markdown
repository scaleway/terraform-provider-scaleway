---
layout: "scaleway"
page_title: "Scaleway: ssh_key"
sidebar_current: "docs-scaleway-resource-ssh_key"
description: |-
  Manages Scaleway user SSH keys.
---

# scaleway_ssh_key

Manages user SSH Keys to access servers provisioned on scaleway.
For additional details please refer to [API documentation](https://developer.scaleway.com/#users-user-get).

## Example Usage

```hcl
resource "scaleway_ssh_key" "test" {
    key = "ssh-rsa <some-key>"
}
```

## Argument Reference

The following arguments are supported:

* `key` - (Required) public key of the SSH key to be added

## Attributes Reference

The following attributes are exported:

* `id` - fingerprint of the SSH key

## Import

Instances can be imported using the `id`, e.g.

```
$ terraform import scaleway_ssh_key.awesome "d1:4c:45:59:a8:ee:e6:41:10:fb:3c:3e:54:98:5b:6f"
```
