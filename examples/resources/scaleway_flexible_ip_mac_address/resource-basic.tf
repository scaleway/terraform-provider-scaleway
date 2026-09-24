### Basic

resource "scaleway_flexible_ip" "main" {}

resource "scaleway_flexible_ip_mac_address" "main" {
  flexible_ip_id = scaleway_flexible_ip.main.id
  type           = "kvm"
}
