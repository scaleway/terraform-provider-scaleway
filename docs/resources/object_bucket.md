---
page_title: "Scaleway: scaleway_object_bucket"
description: |-
  Manages Scaleway object storage buckets.
---

# scaleway_object_bucket

Creates and manages Scaleway object storage buckets. For more information, see [the documentation](https://www.scaleway.com/en/docs/object-storage-feature/).

## Example Usage

```hcl
resource "scaleway_object_bucket" "some_bucket" {
  name = "some-unique-name"
  acl  = "private"
  tags = {
    key = "value"
  }
}
```

## Arguments Reference

The following arguments are supported:

* `name` - (Required) The name of the bucket.
* `tags` - (Optional) A list of tags (key / value) for the bucket.
* `acl` - (Optional) The [canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#canned-acl) you want to apply to the bucket.
* `region` - (Optional) The [region](https://developers.scaleway.com/en/quickstart/#region-definition) in which the bucket should be created.
* `versioning` - (Optional) A state of [versioning](https://docs.aws.amazon.com/AmazonS3/latest/dev/Versioning.html) (documented below)
* `cors_rule` - (Optional) A rule of [Cross-Origin Resource Sharing](https://docs.aws.amazon.com/AmazonS3/latest/dev/cors.html) (documented below).

The `CORS` object supports the following:

* `allowed_headers` (Optional) Specifies which headers are allowed.
* `allowed_methods` (Required) Specifies which methods are allowed. Can be `GET`, `PUT`, `POST`, `DELETE` or `HEAD`.
* `allowed_origins` (Required) Specifies which origins are allowed.
* `expose_headers` (Optional) Specifies expose header in the response.
* `max_age_seconds` (Optional) Specifies time in seconds that browser can cache the response for a preflight request.

The `versioning` object supports the following:

* `enabled` - (Optional) Enable versioning. Once you version-enable a bucket, it can never return to an unversioned state. You can, however, suspend versioning on that bucket.

## Attributes Reference

In addition to all above arguments, the following attribute is exported:

* `id` - The unique name of the bucket.
* `endpoint` - The endpoint URL of the bucket

## Import

Buckets can be imported using the `{region}/{bucketName}` identifier, e.g.

```bash
$ terraform import scaleway_object_bucket.some_bucket fr-par/some-bucket
```
