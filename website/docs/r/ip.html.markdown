---
layout: "scaleway"
page_title: "Scaleway: ip"
sidebar_current: "docs-scaleway-resource-ip"
description: |-
  Manages Scaleway IPs.
---

# scaleway_ip

Provides IPs for servers. This allows IPs to be created, updated and deleted.
For additional details please refer to [API documentation](https://developer.scaleway.com/#ips).

## Example Usage

```hcl
resource "scaleway_ip" "test_ip" {}
```

## Argument Reference

The following arguments are supported:

* `server` - (Optional) ID of server to associate IP with
* `reverse` - (Optional) Reverse DNS of the IP

## Attributes Reference

The following attributes are exported:

* `id` - ID of the new resource
* `ip` - IP of the new resource
* `server` - ID of the associated server resource
* `reverse` - reverse DNS setting of the IP resource

## Import

Instances can be imported using the `id`, e.g.

```
$ terraform import scaleway_ip.jump_host 5faef9cd-ea9b-4a63-9171-9e26bec03dbc
```
