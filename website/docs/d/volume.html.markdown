---
layout: "scaleway"
page_title: "Scaleway: scaleway_volume"
description: |-
  Gets information about a Volume.
---

# scaleway_volume

Gets information about a Volume.

## Example Usage

```hcl
data "scaleway_volume" "data" {
  name = "data"
}

resource "scaleway_server" "test" {
  # ...
}

resource "scaleway_volume_attachment" "data" {
  server = scaleway_server.test.id
  volume = scaleway_volume.data.id
}
```

## Argument Reference

* `name` - (Required) Exact name of the Volume.

## Attributes Reference

`id` is set to the ID of the found Volume. In addition, the following attributes
are exported:


* `size_in_gb` - (Required) size of the volume in GB
* `type` - The type of volume this is, such as `l_ssd`.
* `server` - The ID of the Server which this Volume is currently attached to.
