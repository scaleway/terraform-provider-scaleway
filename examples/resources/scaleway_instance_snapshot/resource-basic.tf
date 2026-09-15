resource "scaleway_instance_snapshot" "main" {
  name      = "some-snapshot-name"
  volume_id = "11111111-1111-1111-1111-111111111111"
}
