### With a reserved IP

resource "scaleway_instance_ip" "ip" {}

resource "scaleway_instance_server" "web" {
  type  = "DEV1-S"
  image = "f974feac-abae-4365-b988-8ec7d1cec10d"

  tags = ["hello", "public"]

  ip_id = scaleway_instance_ip.ip.id
}
