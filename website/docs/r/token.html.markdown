---
layout: "scaleway"
page_title: "Scaleway: token"
sidebar_current: "docs-scaleway-resource-token"
description: |-
  Manages Scaleway Tokens.
---

# scaleway_token

Provides Tokens for scaleway API access. For additional details please refer to [API documentation](https://developer.scaleway.com/#tokens-tokens-post).

## Example Usage

```hcl
resource "scaleway_token" "karls_token" {
    expires = false
    description = "karls scaleway access: karl@company.com"
}
```

## Argument Reference

The following arguments are supported:

* `expires` - (Optional) Define if the token should automatically expire or not
* `email` - (Optional) Scaleway account email. Defaults to registered account
* `password` - (Optional) Scaleway account password. Required for cross-account token management
* `description` - (Optional) Token description

## Attributes Reference

The following attributes are exported:

* `id` - Token ID - can be used to access scaleway API
* `expiration_date` - Expiration date of token, if expiration is requested

## Import

Instances can be imported using the `id`, e.g.

```
$ terraform import scaleway_token.karls_token 5faef9cd-ea9b-4a63-9171-9e26bec03dbc
```
