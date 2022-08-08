---
page_title: "Scaleway: scaleway_object_bucket_policy"
description: |-
Manages Scaleway object storage bucket policy.
---

# scaleway_object_bucket

Creates and manages Scaleway object storage bucket policy.
For more information, see [the documentation](https://www.scaleway.com/en/docs/storage/object/api-cli/using-bucket-policies/).

## Example Usage

```hcl
resource "scaleway_object_bucket_policy" "bucket" {
  name = "some-unique-name"
}

resource "scaleway_object_bucket_policy" "bucket" {
    bucket = scaleway_object_bucket.bucket.name
    policy = jsonencode(
    {
        Id = "MyPolicy"
        Statement = [
        {
            Action = [
                "s3:ListBucket",
               "s3:GetObject",
            ]
           Effect = "Allow"
           Principal = {
               SCW = "*"
            }
           Resource  = [
              "some-unique-name",
              "some-unique-name/*",
            ]
           Sid = "GrantToEveryone"
        },
    ]
    Version = "2012-10-17"
    }
    )
}
```

## Example with aws provider

```hcl
resource "scaleway_object_bucket_policy" "bucket" {
  name = "some-unique-name"
}

resource "scaleway_object_bucket_policy" main {
    bucket = scaleway_object_bucket.bucket.name
    policy = data.aws_iam_policy_document.policy.json
}

data "aws_iam_policy_document" "policy" {
  version = "2012-10-17"
  statement {
    sid = "MyPolicy"
    principals {
      type        = "SCW"
      identifiers = ["project_id:<project_id>"]
    }

    actions = [
      "s3:GetObject",
      "s3:ListBucket",
    ]

    resources = [
      "some-unique-name",
      "some-unique-name/*",
    ]
  }
}
```

## Arguments Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket.
* `policy` - (Required) The policy document. This is a JSON formatted string. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/tutorials/terraform/aws-iam-policy?_ga=2.164714495.1557487853.1659960650-563504983.1635944492).

~> **Important:** The [aws_iam_policy_document](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/data-sources/iam_policy_document) data source may be used, so long as it specifies a principal.

## Attributes Reference

In addition to all above arguments, the following attribute is exported:

* `id` - The unique name of the bucket.
* `region` - The Scaleway region this bucket resides in.

## Import

Buckets can be imported using the `{region}/{bucketName}` identifier, e.g.

```bash
$ terraform import scaleway_object_bucket_policy.some_bucket fr-par/some-bucket
```
