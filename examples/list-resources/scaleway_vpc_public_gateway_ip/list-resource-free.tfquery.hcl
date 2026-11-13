# List only free (unattached) Public Gateway IPs
list "scaleway_vpc_public_gateway_ip" "free" {
  provider = scaleway

  config {
    zones   = ["*"]
    is_free = true
  }
}
