### Example importing a local qcow2 file

resource "scaleway_object_bucket" "bucket" {
  name = "snapshot-qcow-import"
}

resource "scaleway_object" "qcow" {
  bucket = scaleway_object_bucket.bucket.name
  key    = "server.qcow2"
  file   = "myqcow.qcow2"
}

resource "scaleway_instance_snapshot" "snapshot" {
  import {
    bucket = scaleway_object.qcow.bucket
    key    = scaleway_object.qcow.key
  }
}
