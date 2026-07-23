## Example Usage with Grants

resource "scaleway_object_bucket" "main" {
  name = "your-bucket"
}

resource "scaleway_object_bucket_acl" "main" {
  bucket = scaleway_object_bucket.main.id
  access_control_policy {
    grant {
      grantee {
        id   = "<project-id>:<project-id>"
        type = "CanonicalUser"
      }
      permission = "FULL_CONTROL"
    }

    grant {
      grantee {
        id   = "<project-id>"
        type = "CanonicalUser"
      }
      permission = "WRITE"
    }

    owner {
      id = "<project-id>"
    }
  }
}
