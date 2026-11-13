resource "scaleway_secret" "psk" {
  name = "my-s2s-vpn-psk"
}

resource "scaleway_secret_version" "psk" {
  secret_id = scaleway_secret.psk.id
  data      = "your_s2s_vpn.psk"
}

resource "scaleway_s2s_vpn_connection" "main" {
  name                     = "my-connection"
  vpn_gateway_id           = scaleway_s2s_vpn_gateway.main.id
  customer_gateway_id      = scaleway_s2s_vpn_customer_gateway.main.id
  initiation_policy        = "customer_gateway"
  enable_route_propagation = false
  secret_id                = scaleway_secret.psk.id
  secret_version           = scaleway_secret_version.psk.revision

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.scaleway_s2s_vpn_connection_enable_route_propagation.main]
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

action "scaleway_s2s_vpn_connection_enable_route_propagation" "main" {
  config {
    connection_id = scaleway_s2s_vpn_connection.main.id
  }
}
