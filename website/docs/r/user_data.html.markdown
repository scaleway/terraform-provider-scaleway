---
layout: "scaleway"
page_title: "Scaleway: user_data"
sidebar_current: "docs-scaleway-resource-user_data"
description: |-
  Manages Scaleway Server UserData.
---

# scaleway_user_data

Provides user data for servers.
For additional details please refer to [API documentation](https://developer.scaleway.com/#user-data).

## Example Usage

```hcl
resource "scaleway_server" "base" {
  name = "test"
  # ubuntu 14.04
  image = "5faef9cd-ea9b-4a63-9171-9e26bec03dbc"
  type = "C1"
  state = "stopped"
}

resource "scaleway_user_data" "gcp" {
	server = "${scaleway_server.base.id}"
	key = "gcp_username"
	value = "supersecret"
}
```

## Argument Reference

The following arguments are supported:

* `server` - (Required) ID of server to associate the user data with
* `key` - (Required) The key of the user data object
* `value` - (Required) The value of the user data object

## Import

Instances can be imported using the `id`, e.g.

```
$ terraform import scaleway_user_data.gcp userdata-<server-id>-<key>
```
