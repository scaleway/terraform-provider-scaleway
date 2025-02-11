---
subcategory: "Object Storage"
page_title: "Scaleway: scaleway_object_bucket_website_configuration"
---

# Resource: scaleway_object_bucket_website_configuration

The `scaleway_object_bucket_website_configuration` resource allows you to deploy and manage a bucket website with [Scaleway Object storage](https://www.scaleway.com/en/docs/object-storage/).

Refer to the [dedicated documentation](https://www.scaleway.com/en/docs/object-storage/how-to/use-bucket-website/) for more information on bucket websites.

## Example Usage

```terraform
resource "scaleway_object_bucket" "main" {
    name = "MyBucket"
    acl  = "public-read"
}

resource "scaleway_object_bucket_website_configuration" "main" {
    bucket = scaleway_object_bucket.main.id
    index_document {
      suffix = "index.html"
    }
}
```

## Example Usage with a bucket policy

```terraform
resource "scaleway_object_bucket" "main" {
    name = "MyBucket"
    acl  = "public-read"
}

resource "scaleway_object_bucket_policy" "main" {
    bucket = scaleway_object_bucket.main.id
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
    bucket = scaleway_object_bucket.main.id
    index_document {
      suffix = "index.html"
    }
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required, forces new resource) The name of the bucket.

* `index_document` - (Required) The name of the index file for the website [detailed below](#index_document).

* `error_document` - (Optional) The name of the error file for the website [detailed below](#error_document).

* `project_id` - (Defaults to [provider](../index.md#arguments-reference) `project_id`) The ID of the project the bucket is associated with.

~> **Important:** The `project_id` attribute has a particular behavior with s3 products, because the s3 API is scoped by project.
If you are using a project different from the default one, you have to specify the `project_id` for every child resource of the bucket,
like bucket website configurations. Otherwise, Terraform will try to create the child resource with the default project ID and you will get a 403 error.

### index_document

The `index_document` configuration block supports the following arguments:

* `suffix` - (Required) A suffix that is appended to a request targeting a specific directory on the website endpoint.

~> **Important:** The suffix must not be empty and must not include a slash character. The routing is not supported.

## Attributes Reference

The `scaleway_object_bucket_website_configuration` resource exports certain attributes once the bucket website configuration is retrieved. These attributes can be referenced in other parts of your Terraform configuration.

* `id` - The region and bucket name, separated by a slash (e.g. `fr-par/some-bucket`)

* `website_domain` - The domain of the website endpoint. This is used to create DNS alias [records](https://www.scaleway.com/en/docs/network/domains-and-dns/how-to/manage-dns-records/).

* `website_endpoint` - The website endpoint.

~> **Important:** Refer to the [dedicated section on endpoints](https://www.scaleway.com/en/docs/object-storage/concepts/#endpoint) for more information.

## error_document

The error_document configuration block supports the following arguments:

* `key` - (Required) The object key name to use when a 4XX class error occurs.

## Import

Bucket website configurations can be imported using the `{region}/{bucketName}` identifier, as shown below:

```bash
terraform import scaleway_object_bucket_website_configuration.some_bucket fr-par/some-bucket
```

~> **Important:** The `project_id` attribute has a particular behavior with s3 products because the s3 API is scoped by project.
If you are using a project different from the default one, you have to specify the project ID at the end of the import command.

```bash
terraform import scaleway_object_bucket_website_configuration.some_bucket fr-par/some-bucket@xxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxx
```
