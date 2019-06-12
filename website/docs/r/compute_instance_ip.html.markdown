---
layout: "scaleway"
page_title: "Scaleway: scaleway_compute_instance_ip"
sidebar_current: "docs-scaleway-resource-compute-instance-ip"
description: |-
  Manages Scaleway compute instance ip.
---

# scaleway_compute_instance_ip

Creates and manages Scaleway compute instance IPs.

## Example Usage

```hcl
resource "scaleway_compute_instance_ip" "test" {}
```

## Argument Reference

The following arguments are supported:

* `project_id` - (Optional) The id of the project you want to attach this resource to.
* `reverse` - (Optional) The reverse dns for this IP.
* `zone` - (Optional) The zone you want to attach this resource to.

## Attributes Reference

The following attributes are exported:

* `project_id` - The id of the project your resource is attached to.
* `reverse` - The reverse dns for this IP.
* `server_id` - The id of the server this resource is attached to.
* `zone` - The zone your resource is attached to.


## Import

Instances can be imported using the `zone/id`, e.g.

```
$ terraform import scaleway_compute_instance_ip.test fr-par-1/11111111-1111-1111-1111-111111111111
```
