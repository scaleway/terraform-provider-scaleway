---
page_title: "Scaleway: scaleway_object_bucket_website_configuration"
description: |-
Manages Scaleway website on object storage buckets.
---

# scaleway_object_bucket_website_configuration

Provides an Object bucket website configuration resource.
For more information, see [Hosting Websites on Object bucket](https://www.scaleway.com/en/docs/storage/object/how-to/use-bucket-website/).

## Example Usage

```hcl
resource "scaleway_object_bucket" "main" {
    name = "MyBucket"
    acl  = "public-read"
}

resource "scaleway_object_bucket_website_configuration" "main" {
    bucket = scaleway_object_bucket.main.name
    index_document {
      suffix = "index.html"
    }
}
```

## Example with `policy`

```hcl
resource "scaleway_object_bucket" "main" {
    name = "MyBucket"
    acl  = "public-read"
}

resource "scaleway_object_bucket_policy" "main" {
    bucket = scaleway_object_bucket.main.name
    policy = jsonencode(
    {
        "Version" = "2012-10-17",
        "Id" = "MyPolicy",
        "Statement" = [
        {
           "Sid" = "GrantToEveryone",
           "Effect" = "Allow",
           "Principal" = "*",
           "Action" = [
              "s3:GetObject"
           ],
           "Resource":[
              "<bucket-name>/*"
           ]
        }
        ]
    })
}

resource "scaleway_object_bucket_website_configuration" "main" {
    bucket = scaleway_object_bucket.main.name
    index_document {
      suffix = "index.html"
    }
}
```

## Attributes Reference

The following arguments are supported:

* `bucket` - (Required, Forces new resource) The name of the bucket.
* `index_document` - (Required) The name of the index document for the website [detailed below](#index_document).
* `error_document` - (Optional) The name of the error document for the website [detailed below](#error_document).

## index_document

The `index_document` configuration block supports the following arguments:

* `suffix` - (Required) A suffix that is appended to a request that is for a directory on the website endpoint.

~> **Important:** The suffix must not be empty and must not include a slash character. The routing is not supported.

In addition to all above arguments, the following attribute is exported:

* `id` - The bucket and region separated by a slash (/)
* `website_domain` - The domain of the website endpoint. This is used to create DNS alias [records](https://www.scaleway.com/en/docs/network/dns-cloud/how-to/manage-dns-records).
* `website_endpoint` - The website endpoint.

~> **Important:** Please check our concepts section to know more about the [endpoint](https://www.scaleway.com/en/docs/storage/object/concepts/#endpoint).

## error_document

The error_document configuration block supports the following arguments:

* `key` - (Required) The object key name to use when a 4XX class error occurs.

## Import

Website configuration Bucket can be imported using the `{region}/{bucketName}` identifier, e.g.

```bash
$ terraform import scaleway_object_bucket_website_configuration.some_bucket fr-par/some-bucket
```
