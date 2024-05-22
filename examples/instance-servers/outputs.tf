output "ips_v4" {
  description = "The public IPv4 addresses of the created instance servers"
  value       = scaleway_instance_server.server[*].public_ip
}
