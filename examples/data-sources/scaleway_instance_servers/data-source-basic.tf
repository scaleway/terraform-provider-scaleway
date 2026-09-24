### Basic

# Find servers by tag
data "scaleway_instance_servers" "my_key" {
  tags = ["tag"]
}

# Find servers by name and zone
data "scaleway_instance_servers" "my_key" {
  name = "myserver"
  zone = "fr-par-2"
}
