### Example with AWS provider

# AWS provider configuration (with Scaleway credentials)
provider "aws" {
  shared_config_files      = ["/home/user/.aws/config"]
  shared_credentials_files = ["/home/user/.aws/credentials"]
  profile                  = "aws-profile"

  skip_region_validation      = true
  skip_credentials_validation = true
  skip_requesting_account_id  = true
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
