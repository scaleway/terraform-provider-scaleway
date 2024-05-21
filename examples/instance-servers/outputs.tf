output "ips_v4" {
  value = scaleway_instance_server.server[*].public_ip
}