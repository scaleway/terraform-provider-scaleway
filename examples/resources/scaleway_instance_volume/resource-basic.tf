resource "scaleway_instance_volume" "server_volume" {
  type       = "l_ssd"
  name       = "some-volume-name"
  size_in_gb = 20
}
