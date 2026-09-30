### Basic

variable "user_data" {
  type = map(any)
  default = {
    "cloud-init" = <<-EOF
    #cloud-config
    apt-update: true
    apt-upgrade: true
    EOF
    "foo"        = "bar"
  }
}

# User data with a single value
resource "scaleway_instance_user_data" "main" {
  server_id = scaleway_instance_server.main.id
  key       = "foo"
  value     = "bar"
}

# User Data with many keys.
resource "scaleway_instance_user_data" "data" {
  server_id = scaleway_instance_server.main.id
  for_each  = var.user_data
  key       = each.key
  value     = each.value
}

resource "scaleway_instance_server" "main" {
  image = "ubuntu_focal"
  type  = "DEV1-S"
}
