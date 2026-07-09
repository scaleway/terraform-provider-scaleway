### Example Usage with an IAM user

# Project ID
data "scaleway_account_project" "default" {
  name = "default"
}

# IAM configuration
data "scaleway_iam_user" "user" {
  email = "user@scaleway.com"
}
resource "scaleway_iam_policy" "policy" {
  name    = "object-storage-policy"
  user_id = data.scaleway_iam_user.user.id
  rule {
    project_ids          = [data.scaleway_account_project.default.id]
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
    Version = "2023-04-17",
    Id      = "MyBucketPolicy",
    Statement = [
      {
        Effect    = "Allow"
        Action    = ["s3:*"]
        Principal = { SCW = "user_id:${data.scaleway_iam_user.user.id}" }
        Resource = [
          scaleway_object_bucket.bucket.name,
          "${scaleway_object_bucket.bucket.name}/*",
        ]
      },
    ]
  })
}
