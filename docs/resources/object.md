---
page_title: "Scaleway: scaleway_object"
description: |-
Manages Scaleway object storage objects.
---

# scaleway_object

Creates and manages Scaleway object storage objects.
For more information, see [the documentation](https://www.scaleway.com/en/docs/object-storage-feature/).

## Example Usage

```hcl
resource "scaleway_object_bucket" "some_bucket" {
  name = "some-unique-name"
}

resource scaleway_object "some_file" {
  bucket = scaleway_object_bucket.some_bucket.name
  key = "object_path"
  
  file = "myfile"
  hash = filemd5("myfile")
}
```

## Arguments Reference


The following arguments are supported:

* `bucket` - (Required) The name of the bucket.
* `key` - (Required) The path of the object.
* `file` - (Optional) The name of the file to upload, defaults to an empty file
* `hash` - (Optional) Hash of the file, used to trigger upload on file change
* `storage_class` - (Optional) Specifies the Scaleway [storage class](https://www.scaleway.com/en/docs/storage/object/concepts/#storage-class) `STANDARD`, `GLACIER`, `ONEZONE_IA` used to store the object.
* `visibility` - (Optional) Visibility of the object, `public-read` or `private`
* `metadata` - (Optional) Map of metadata used for the object, keys must be lowercase
* `tags` - (Optional) Map of tags

## Attributes Reference

In addition to all above arguments, the following attribute is exported:

* `id` - The path of the object, including bucket name.
* `region` - The Scaleway region this bucket resides in.

## Import

Objects can be imported using the `{region}/{bucketName}/{objectKey}` identifier, e.g.

```bash
$ terraform import scaleway_object.some_object fr-par/some-bucket/some-file
```
