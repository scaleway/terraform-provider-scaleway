---
page_title: "Scaleway: scaleway_object_bucket_acl"
description: |-
Manages Scaleway object storage bucket ACL resource.
---

# scaleway_object_bucket

Creates and manages Scaleway object storage bucket ACL.
For more information, see [the documentation](https://www.scaleway.com/en/docs/storage/object/concepts/#access-control-list-(acl)).

-> **Note:** `terraform destroy`  does not delete the Object Bucket ACL but does remove the resource from Terraform state.

-> **Note:** [Account identifiers](https://docs.aws.amazon.com/general/latest/gr/acct-identifiers.html) is not supported by scaleway.

## Example Usage

```hcl
resource "scaleway_object_bucket" "some_bucket" {
  name = "some-unique-name"
}

resource "scaleway_object_bucket_acl" "main" {
  bucket = scaleway_object_bucket.main.name
  acl = "private"
}
```

## Arguments Reference


The following arguments are supported:

* `bucket` - (Required) The name of the bucket.
* `acl` - (Optional) The canned ACL you want to apply to the bucket.
* `region` - (Optional) The [region](https://developers.scaleway.com/en/quickstart/#region-definition) in which the bucket should be created.

## The ACL

Please check the [canned ACL](https://docs.aws.amazon.com/AmazonS3/latest/userguide/acl_overview.html#canned-acl)

## Attributes Reference

In addition to all above arguments, the following attribute is exported:

* `id` - The `region`,`bucket` and `acl` separated by (`/`).

## Import

Buckets can be imported using the `{region}/{bucketName}/{acl}` identifier, e.g.

```bash
$ terraform import scaleway_object_bucket_acl.some_bucket fr-par/some-bucket
/private```
