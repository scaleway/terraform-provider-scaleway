### Basic

resource "scaleway_mongodb_snapshot" "main" {
  instance_id = scaleway_mongodb_instance.main.id
  name        = "name-snapshot"
  expires_at  = "2024-12-31T23:59:59Z"
}
