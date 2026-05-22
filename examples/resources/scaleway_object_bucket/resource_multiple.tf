resource "scaleway_object_bucket" "main" {
  name   = "mybuckectid"
  region = "fr-par"

  # This lifecycle configuration rule will make that all objects that got a filter key that start with (path1/) be transferred
  # from their default storage class (STANDARD, ONEZONE_IA) to GLACIER after 120 days counting
  # from their creation and then 365 days after that they will be expired and deleted.
  lifecycle_rule {
    id      = "id1"
    prefix  = "path1/"
    enabled = true

    expiration {
      days = 365
    }

    transition {
      days          = 120
      storage_class = "GLACIER"
    }
  }

  # This lifecycle configuration rule specifies that all objects (identified by the key name prefix (path2/) in the rule)
  # from their creation and then 50 days after that they will be expired and deleted.
  lifecycle_rule {
    id      = "id2"
    prefix  = "path2/"
    enabled = true

    expiration {
      days = "50"
    }
  }

  # This lifecycle configuration rule remove any object with (path3/) prefix that match
  # with the tags one day after creation.
  lifecycle_rule {
    id      = "id3"
    prefix  = "path3/"
    enabled = false

    tags = {
      "tagKey"    = "tagValue"
      "terraform" = "hashicorp"
    }

    expiration {
      days = "1"
    }
  }

  # This lifecycle configuration rule specifies a tag-based filter (tag1/value1).
  # This rule directs Scaleway S3 to transition objects S3 Glacier class soon after creation.
  lifecycle_rule {
    id      = "id4"
    enabled = true

    tags = {
      "tag1" = "value1"
    }

    transition {
      days          = 1
      storage_class = "GLACIER"
    }
  }

  # This lifecycle configuration rule specifies with the AbortIncompleteMultipartUpload action to
  # stop incomplete multipart uploads (identified by the key name prefix (path5/) in the rule)
  # if they aren't completed within a specified number of days after initiation.
  # Note: It's not recommended using prefix/ for AbortIncompleteMultipartUpload as any incomplete multipart upload will be billed
  lifecycle_rule {
    #  prefix  = "path5/"
    enabled                                = true
    abort_incomplete_multipart_upload_days = 30
  }
}
