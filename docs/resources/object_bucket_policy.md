---
subcategory: "Object Storage"
page_title: "Scaleway: scaleway_object_bucket_policy"
---

# Resource: scaleway_object_bucket_policy

Creates and manages Scaleway object storage bucket policy.
For more information, see [the documentation](https://www.scaleway.com/en/docs/storage/object/api-cli/bucket-policy/).

## Example Usage

### Example with an IAM user

```hcl
# Project ID
data "scaleway_account_project" "default" {
  name = "default"
}

# IAM configuration
data "scaleway_iam_user" "user" {
  email = "user@scaleway.com"
}
resource "scaleway_iam_policy" "policy" {
  name = "object-storage-policy"
  user_id = data.scaleway_iam_user.user.id
  rule {
    project_ids = [data.scaleway_account_project.default.id]
    permission_set_names = ["ObjectStorageFullAccess"]
  }
}

# Object storage configuration
resource "scaleway_object_bucket" "bucket" {
  name = "some-unique-name"
}
resource "scaleway_object_bucket_policy" "policy" {
  bucket = scaleway_object_bucket.bucket.name
  policy = jsonencode({
    Version   = "2023-04-17",
    Id      = "MyBucketPolicy",
    Statement = [
      {
        Effect = "Allow"
        Action = ["s3:*"]
        Principal = { SCW = "user_id:${data.scaleway_iam_user.user.id}" }
        Resource  = [
          scaleway_object_bucket.bucket.name,
          "${scaleway_object_bucket.bucket.name}/*",
        ]
      },
    ]
  })
}
```

### Example with an IAM application

#### Creating a bucket and delegating read access to an application

```hcl
# Project ID
data "scaleway_account_project" "default" {
  name = "default"
}

# IAM configuration
resource "scaleway_iam_application" "reading-app" {
  name = "reading-app"
}
resource "scaleway_iam_policy" "policy" {
  name = "object-storage-policy"
  application_id = scaleway_iam_application.reading-app.id
  rule {
    project_ids = [data.scaleway_account_project.default.id]
    permission_set_names = ["ObjectStorageBucketsRead"]
  }
}

# Object storage configuration
resource "scaleway_object_bucket" "bucket" {
  name = "some-unique-name"
}
resource "scaleway_object_bucket_policy" "policy" {
  bucket = scaleway_object_bucket.bucket.id
  policy = jsonencode(
    {
      Version = "2023-04-17",
      Statement = [
        {
          Sid    = "Delegate read access",
          Effect = "Allow",
          Principal = {
            SCW = "application_id:${scaleway_iam_application.reading-app.id}"
          },
          Action = [
            "s3:ListBucket",
            "s3:GetObject",
          ]
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

#### Reading the bucket with the application

```hcl
data "scaleway_iam_application" "reading-app" {
  name = "reading-app"
}
resource "scaleway_iam_api_key" "reading-api-key" {
  application_id = data.scaleway_iam_application.reading-app.id
}

provider "scaleway" {
  access_key = scaleway_iam_api_key.reading-api-key.access_key
  secret_key = scaleway_iam_api_key.reading-api-key.secret_key
  alias = "reading-profile"
}

data scaleway_object_bucket bucket {
  provider = scaleway.reading-profile
  name = "some-unique-name"
  depends_on = [scaleway_iam_api_key.reading-api-key]
}
```

### Example with AWS provider

```hcl
# AWS provider configuration (with Scaleway credentials)
provider "aws" {
  shared_config_files      = ["/home/user/.aws/config"]
  shared_credentials_files = ["/home/user/.aws/credentials"]
  profile                  = "aws-profile"

  skip_region_validation = true
  skip_credentials_validation = true
  skip_requesting_account_id = true
}

# Scaleway project ID
data "scaleway_account_project" "default" {
  name = "default"
}

# Object storage configuration
resource "scaleway_object_bucket" "bucket" {
  name = "some-unique-name"
}
resource "scaleway_object_bucket_policy" "main" {
  bucket = scaleway_object_bucket.bucket.id
  policy = data.aws_iam_policy_document.policy.json
}

# AWS data source
data "aws_iam_policy_document" "policy" {
  version = "2012-10-17"
  statement {
    sid    = "Delegate access"
    effect = "Allow"
    principals {
      type        = "SCW"
      identifiers = ["project_id:${data.scaleway_account_project.default.id}"]
    }
    actions = ["s3:ListBucket"]
    resources = [
      "${scaleway_object_bucket.bucket.name}",
      "${scaleway_object_bucket.bucket.name}/*"
    ]
  }
}
```

### Example with deprecated version 2012-10-17

```hcl
# Project ID
data "scaleway_account_project" "default" {
  name = "default"
}

# Object storage configuration
resource "scaleway_object_bucket" "bucket" {
  name   = "mia-cross-crash-tests"
  region = "fr-par"
}
resource "scaleway_object_bucket_policy" "policy" {
  bucket = scaleway_object_bucket.bucket.name
  policy = jsonencode({
    Version   = "2012-10-17",
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:ListBucket",
          "s3:GetObjectTagging"
        ]
        Principal = { SCW = "project_id:${data.scaleway_account_project.default.id}" }
        Resource  = [
          scaleway_object_bucket.bucket.name,
          "${scaleway_object_bucket.bucket.name}/*",
        ]
      },
    ]
  })
}
```

**NB:** To configure the AWS provider with Scaleway credentials, please visit this [tutorial](https://www.scaleway.com/en/docs/storage/object/api-cli/object-storage-aws-cli/).

## Arguments Reference

The following arguments are supported:

* `bucket` - (Required) The name of the bucket, or its Terraform ID.
* `policy` - (Required) The policy document. This is a JSON formatted string. For more information about building AWS IAM policy documents with Terraform, see the [AWS IAM Policy Document Guide](https://learn.hashicorp.com/tutorials/terraform/aws-iam-policy?_ga=2.164714495.1557487853.1659960650-563504983.1635944492).
* `project_id` - (Defaults to [provider](../index.md#arguments-reference) `project_id`) The ID of the project the bucket is associated with.

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
