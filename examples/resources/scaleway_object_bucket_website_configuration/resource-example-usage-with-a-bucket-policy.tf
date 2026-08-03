## Example Usage with a bucket policy

resource "scaleway_object_bucket" "main" {
  name = "MyBucket"
  acl  = "public-read"
}

resource "scaleway_object_bucket_policy" "main" {
  bucket = scaleway_object_bucket.main.id
  policy = jsonencode(
    {
      "Version" = "2012-10-17",
      "Id"      = "MyPolicy",
      "Statement" = [
        {
          "Sid"       = "GrantToEveryone",
          "Effect"    = "Allow",
          "Principal" = "*",
          "Action" = [
            "s3:GetObject"
          ],
          "Resource" : [
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
