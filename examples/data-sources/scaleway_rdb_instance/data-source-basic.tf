## Example Usage

# Get info by name
data "scaleway_rdb_instance" "my_instance" {
  name = "foobar"
}

# Get info by instance ID
data "scaleway_rdb_instance" "my_instance" {
  instance_id = "11111111-1111-1111-1111-111111111111"
}

# Get other attributes
output "load_balancer_ip_addr" {
  description = "IP address of load balancer"
  value       = data.scaleway_rdb_instance.my_instance.load_balancer.0.ip
}
