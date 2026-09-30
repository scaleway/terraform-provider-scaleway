### How to export to Object Storage

resource "scaleway_object_bucket" "my-import-bucket" {
  name = "snapshot-bucket-to-import"
}

resource "scaleway_object" "qcow-object" {
  bucket = scaleway_object_bucket.snapshot-bucket.name
  key    = "export/my-snapshot.qcow2"
}

resource "scaleway_block_volume" "to_export" {
  iops = 5000
  name = "to-export"

  export {
    bucket = "snapshot-bucket-to-import"
    key    = "exports/my-snapshot.qcow2"
  }
}
