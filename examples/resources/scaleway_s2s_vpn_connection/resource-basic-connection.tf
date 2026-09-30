### Basic Connection

resource "scaleway_vpc" "vpc" {
  name = "my-vpc"
}

resource "scaleway_vpc_private_network" "pn" {
  name   = "my-private-network"
  vpc_id = scaleway_vpc.vpc.id
  ipv4_subnet {
    subnet = "10.0.1.0/24"
  }
}

resource "scaleway_s2s_vpn_gateway" "gateway" {
  name               = "my-vpn-gateway"
  gateway_type       = "VGW-S"
  private_network_id = scaleway_vpc_private_network.pn.id
}

resource "scaleway_s2s_vpn_customer_gateway" "customer_gw" {
  name        = "my-customer-gateway"
  ipv4_public = "203.0.113.1"
  asn         = 65000
}

resource "scaleway_s2s_vpn_routing_policy" "policy" {
  name              = "my-routing-policy"
  prefix_filter_in  = ["10.0.2.0/24"]
  prefix_filter_out = ["10.0.1.0/24"]
}

resource "scaleway_s2s_vpn_connection" "main" {
  name                     = "my-vpn-connection"
  vpn_gateway_id           = scaleway_s2s_vpn_gateway.gateway.id
  customer_gateway_id      = scaleway_s2s_vpn_customer_gateway.customer_gw.id
  initiation_policy        = "customer_gateway"
  enable_route_propagation = true

  bgp_config_ipv4 {
    routing_policy_id = scaleway_s2s_vpn_routing_policy.policy.id
    private_ip        = "169.254.0.1/30"
    peer_private_ip   = "169.254.0.2/30"
  }

  ikev2_ciphers {
    encryption = "aes256"
    integrity  = "sha256"
    dh_group   = "modp2048"
  }

  esp_ciphers {
    encryption = "aes256"
    integrity  = "sha256"
    dh_group   = "modp2048"
  }
}
