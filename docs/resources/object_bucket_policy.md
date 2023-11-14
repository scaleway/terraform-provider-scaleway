---
subcategory: "Object Storage"
page_title: "Scaleway: scaleway_object_bucket_policy"
---

# scaleway_object_bucket

Creates and manages Scaleway object storage bucket policy.
For more information, see [the documentation](https://www.scaleway.com/en/docs/storage/object/api-cli/bucket-policy/).

## Example Usage

```hcl
resource "scaleway_object_bucket" "bucket" {
  name = "some-unique-name"
}

resource "scaleway_iam_application" "main" {
  name        = "My application"
  description = "a description"
}

resource "scaleway_object_bucket_policy" "policy" {
  bucket = scaleway_object_bucket.bucket.id
  policy = jsonencode(
    {
      Version = "2023-04-17",
      Id      = "MyBucketPolicy",
      Statement = [
        {
          Sid    = "Delegate access",
          Effect = "Allow",
          Principal = {
            SCW = "application_id:${scaleway_iam_application.main.id}"
          },
          Action = "s3:ListBucket",
          Resource = [
            "${scaleway_object_bucket.bucket.name}",
            "${scaleway_object_bucket.bucket.name}/*"
          ]
        }
      ]
    }
  )
}
```

## Example with aws provider

```hcl
resource "scaleway_object_bucket" "bucket" {
  name = "some-unique-name"
}

resource "scaleway_object_bucket_policy" "main" {
  bucket = scaleway_object_bucket.bucket.id
  policy = data.aws_iam_policy_document.policy.json
}

data "aws_iam_policy_document" "policy" {
  version = "2023-04-17"
  id      = "MyBucketPolicy"

  statement {
    sid    = "Delegate access"
    effect = "Allow"

    principals {
      type        = "SCW"
      identifiers = ["application_id:<APPLICATION_ID>"]
    }

    actions = ["s3:ListBucket"]

    resources = [
      "${scaleway_object_bucket.bucket.name}",
      "${scaleway_object_bucket.bucket.name}/*"
    ]
  }
}
```

## Arguments Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket.
* `policy` - (Required) The policy document. This is a JSON formatted string. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/tutorials/terraform/aws-iam-policy?_ga=2.164714495.1557487853.1659960650-563504983.1635944492).
* `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the bucket is associated with.

~> **Important:** The [aws_iam_policy_document](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) data source may be used, so long as it specifies a principal.

## Attributes Reference

In addition to all above arguments, the following attribute is exported:

* `id` - The ID of the policy, which is the ID of the bucket.

~> **Important:** Object buckets' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{name}`, e.g. `fr-par/bucket-name`

* `region` - The Scaleway region this bucket resides in.

## Import

Buckets can be imported using the `{region}/{bucketName}` identifier, e.g.

```bash
$ terraform import scaleway_object_bucket_policy.some_bucket fr-par/some-bucket
```
