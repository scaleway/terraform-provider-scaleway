Export a block snapshot to Scaleway Object Storage.

This action exports a block snapshot to a specified bucket in Scaleway Object Storage. The snapshot will be saved as a QCOW2 file with the specified key.

-> **Note:** This operation may take some time depending on the size of the snapshot. You can use the `wait` parameter to wait for the operation to complete.

## Example Usage

```hcl
resource "scaleway_block_snapshot" "example" {
  name      = "example-snapshot"
  volume_id = scaleway_block_volume.example.id
}

resource "scaleway_object_bucket" "example" {
  name = "example-bucket"
  region = "fr-par"
}

resource "scaleway_block_export_snapshot" "example" {
  snapshot_id = scaleway_block_snapshot.example.id
  bucket      = scaleway_object_bucket.example.name
  key         = "snapshots/example-snapshot.qcow2"
  wait        = true
}
```

## Argument Reference

- `snapshot_id` - (Required) The ID of the block snapshot to export.
- `zone` - (Optional) The zone where the snapshot is located. If not specified, it will be extracted from the snapshot ID.
- `bucket` - (Required) The name of the bucket where the snapshot will be exported.
- `key` - (Required) The object key (path) where the snapshot will be saved in the bucket.
- `wait` - (Optional) Whether to wait for the export operation to complete. Defaults to `false`.