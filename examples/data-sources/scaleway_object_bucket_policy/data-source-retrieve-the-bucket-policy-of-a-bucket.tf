## Retrieve the bucket policy of a bucket

data "scaleway_object_bucket_policy" "main" {
  bucket = "bucket.test.com"
}
