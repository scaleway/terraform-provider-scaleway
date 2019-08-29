---
layout: "scaleway"
page_title: "Scaleway: bucket"
description: |-
  Manages Scaleway buckets.
---

# scaleway_bucket

**DEPRECATED**: This resource is deprecated and will be removed in `v2.0+`.
Please use `scaleway_object_bucket` instead.

Creates Scaleway object storage buckets.

## Example Usage

```hcl
resource "scaleway_bucket" "test" {
  name = "sample-bucket"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name of the Scaleway objectstorage bucket

## Attributes Reference

The following attributes are exported:

* `name` - Name of the resource

## Import

Instances can be imported using the `name`, e.g.

```
$ terraform import scaleway_bucket.releases releases
```
