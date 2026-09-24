#### Creating a bucket and delegating read access to an application

# Project ID
data "scaleway_account_project" "default" {
  name = "default"
}

# IAM configuration
resource "scaleway_iam_application" "reading-app" {
  name = "reading-app"
}
resource "scaleway_iam_policy" "policy" {
  name           = "object-storage-policy"
  application_id = scaleway_iam_application.reading-app.id
  rule {
    project_ids          = [data.scaleway_account_project.default.id]
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
