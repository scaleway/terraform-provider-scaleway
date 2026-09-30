### Example with deprecated version 2012-10-17

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
    Version = "2012-10-17",
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:ListBucket",
          "s3:GetObjectTagging"
        ]
        Principal = { SCW = "project_id:${data.scaleway_account_project.default.id}" }
        Resource = [
          scaleway_object_bucket.bucket.name,
          "${scaleway_object_bucket.bucket.name}/*",
        ]
      },
    ]
  })
}
