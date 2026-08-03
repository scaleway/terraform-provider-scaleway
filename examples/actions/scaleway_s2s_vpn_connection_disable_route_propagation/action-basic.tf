resource "scaleway_vpc" "main" {
  name = "tf-test-vpc-disable-route-prop"
}

resource "scaleway_vpc_private_network" "main" {
  vpc_id = scaleway_vpc.main.id
  ipv4_subnet {
    subnet = "10.0.0.0/24"
  }
}

resource "scaleway_instance_ip" "customer_ip" {}

resource "scaleway_s2s_vpn_gateway" "main" {
  name               = "tf-test-vpn-gateway-disable-route-prop"
  gateway_type       = "VGW-S"
  private_network_id = scaleway_vpc_private_network.main.id
  region             = "fr-par"
  zone               = "fr-par-1"
}

resource "scaleway_s2s_vpn_customer_gateway" "main" {
  name        = "tf-test-customer-gateway-disable-route-prop"
  ipv4_public = scaleway_instance_ip.customer_ip.address
  asn         = 65000
  region      = "fr-par"
}

resource "scaleway_s2s_vpn_routing_policy" "main" {
  name              = "tf-test-routing-policy-disable-route-prop"
  prefix_filter_in  = ["10.0.1.0/24"]
  prefix_filter_out = ["10.0.0.0/24"]
  region            = "fr-par"
}

resource "scaleway_s2s_vpn_connection" "main" {
  name                     = "tf-test-connection-disable-route-prop"
  vpn_gateway_id           = scaleway_s2s_vpn_gateway.main.id
  customer_gateway_id      = scaleway_s2s_vpn_customer_gateway.main.id
  initiation_policy        = "customer_gateway"
  enable_route_propagation = true
  region                   = "fr-par"

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.scaleway_s2s_vpn_connection_disable_route_propagation.main]
    }
  }

  bgp_config_ipv4 {
    routing_policy_id = scaleway_s2s_vpn_routing_policy.main.id
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

action "scaleway_s2s_vpn_connection_disable_route_propagation" "main" {
  config {
    connection_id = scaleway_s2s_vpn_connection.main.id
    region        = "fr-par"
  }
}
