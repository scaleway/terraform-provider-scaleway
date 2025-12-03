---
subcategory: "Object Storage"
page_title: "Scaleway: scaleway_object"
---

# scaleway_object

The `scaleway_object` data source is used to retrieve information about an Object Storage object.

Refer to the Object Storage [documentation](https://www.scaleway.com/en/docs/object-storage/how-to/create-a-bucket/) for more information.

## Retrieve an Object Storage object

The following example demonstrates how to retrieve metadata about an object stored in a bucket:

```hcl
resource "scaleway_object_bucket" "main" {
  name = "bucket.test.com"
}

resource "scaleway_object" "example" {
  bucket = scaleway_object_bucket.main.name
  key    = "example.txt"
  content = "Hello world!"
}

data "scaleway_object" "selected" {
  bucket = scaleway_object.example.bucket
  key    = scaleway_object.example.key
}
```

## Argument Reference

This section lists the arguments that you can provide to the `scaleway_object` data source to filter and retrieve the desired Object Storage bucket. Each argument has a specific purpose:

- `bucket` - (Required) The name of the bucket, or its terraform ID (`{region}/{name}`)
- `key` - (Required) The key (path or filename) of the object within the bucket.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#zones) in which the bucket exists.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project with which the bucket is associated.

## Attributes Reference

The `scaleway_object` data source exports certain attributes once the object information is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

In addition to all above arguments, the following attribute is exported:

* `id` - The unique identifier of the object.

~> **Important**: Object IDs are regional, and follow the format {region}/{bucket}/{key}, e.g. fr-par/bucket-name/example.txt.
