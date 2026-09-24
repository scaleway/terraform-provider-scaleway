# Get info by snapshot name
data "scaleway_instance_snapshot" "by_name" {
  name = "my-snapshot-name"
}

# Get info by snapshot ID
data "scaleway_instance_snapshot" "by_id" {
  snapshot_id = "11111111-1111-1111-1111-111111111111"
}
