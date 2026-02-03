The `scaleway_s2s_vpn_connection_enable_route_propagation` action enables route propagation on an S2S VPN connection. This allows all allowed prefixes (defined in a routing policy) to be announced in the BGP session, so that traffic can flow between the attached VPC and the on-premises infrastructure along the announced routes.

Note that by default, even when route propagation is enabled, all routes are blocked. It is essential to attach a routing policy to the connection to define the ranges of routes to announce.

Refer to the [S2S VPN documentation](https://www.scaleway.com/en/docs/network/s2s-vpn/) and [API documentation](https://www.scaleway.com/en/developers/api/s2s-vpn/) for more information.

## Example Usage

```terraform
resource "scaleway_s2s_vpn_connection" "main" {
  name                     = "my-connection"
  vpn_gateway_id           = scaleway_s2s_vpn_gateway.main.id
  customer_gateway_id      = scaleway_s2s_vpn_customer_gateway.main.id
  initiation_policy        = "customer_gateway"
  enable_route_propagation = false

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
```
