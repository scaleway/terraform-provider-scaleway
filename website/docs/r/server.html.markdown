---
layout: "scaleway"
page_title: "Scaleway: server"
sidebar_current: "docs-scaleway-resource-server"
description: |-
  Manages Scaleway servers.
---

# scaleway_server

Provides servers. This allows servers to be created, updated and deleted.
For additional details please refer to [API documentation](https://developer.scaleway.com/#servers).

## Example Usage

```hcl
resource "scaleway_server" "test" {
  name  = "test"
  image = "5faef9cd-ea9b-4a63-9171-9e26bec03dbc"
  type  = "VC1M"

  volume {
    size_in_gb = 20
    type       = "l_ssd"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) name of server
* `image` - (Required) base image of server
* `type` - (Required) type of server
* `bootscript` - (Optional) server bootscript
* `boot_type` - (Optional) the boot mechanism for this server. Possible values include `local` and `bootscript`
* `tags` - (Optional) list of tags for server
* `enable_ipv6` - (Optional) enable ipv6
* `dynamic_ip_required` - (Optional) make server publicly available
* `public_ip` - (Optional) set a public ip previously created (a real ip is expected here, not its resource id)
* `security_group` - (Optional) assign security group to server
* `volume` - (Optional) attach additional volumes to your instance (see below)
* `public_ipv6` - (Read Only) if `enable_ipv6` is set this contains the ipv6 address of your instance
* `state` - (Optional) allows you to define the desired state of your server. Valid values include (`stopped`, `running`)
* `state_detail` - (Read Only) contains details from the scaleway API the state of your instance

Field `name`, `type`, `tags`, `dynamic_ip_required`, `security_group` are editable.

## Volume

You can attach additional volumes to your instance, which will share the lifetime
of your `scaleway_server` resource.

~> **Warning:** Using the `volume` attribute does not modify the System Volume provided default with every `scaleway_server` instance. Instead it adds additional volumes to the server instance.

~> **Warning:** Some instance types require an additional volume to work. This includes for example *START-1M* and *VC1M*. If you run into this issue add an additional volume of the specified size.

The `volume` mapping supports the following:

* `type` - (Required) The type of volume. Can be `"l_ssd"`
* `size_in_gb` - (Required) The size of the volume in gigabytes.


## Attributes Reference

The following attributes are exported:

* `id` - id of the new resource
* `private_ip` - private ip of the new resource
* `public_ip` - public ip of the new resource

## Import

Instances can be imported using the `id`, e.g.

```
$ terraform import scaleway_server.web 5faef9cd-ea9b-4a63-9171-9e26bec03dbc
```
