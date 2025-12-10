---
subcategory: "S2S VPN"
page_title: "Scaleway: scaleway_s2s_vpn_connection"
---

# Resource: scaleway_s2s_vpn_connection

Creates and manages Scaleway Site-to-Site VPN Connections.
A connection links a Scaleway VPN Gateway to a Customer Gateway and establishes an IPSec tunnel with BGP routing.

For more information, see [the main documentation](https://www.scaleway.com/en/docs/site-to-site-vpn/reference-content/understanding-s2svpn/).

## Example Usage

### Basic Connection

```terraform
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
  name                = "my-vpn-connection"
  vpn_gateway_id      = scaleway_s2s_vpn_gateway.gateway.id
  customer_gateway_id = scaleway_s2s_vpn_customer_gateway.customer_gw.id
  initiation_policy   = "customer_gateway"
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
```

## Argument Reference

The following arguments are supported:

- `vpn_gateway_id` - (Required) The ID of the VPN gateway to attach to the connection.
- `customer_gateway_id` - (Required) The ID of the customer gateway to attach to the connection.
- `initiation_policy` - (Optional) Defines who initiates the IPSec tunnel.
- `enable_route_propagation` - (Optional) Defines whether route propagation is enabled or not.
- `bgp_config_ipv4` - (Optional) BGP configuration for IPv4. See [BGP Config](#bgp-config) below.
- `bgp_config_ipv6` - (Optional) BGP configuration for IPv6. See [BGP Config](#bgp-config) below.
- `ikev2_ciphers` - (Optional) IKEv2 cipher configuration for Phase 1 (tunnel establishment). See [Cipher Config](#cipher-config) below.
- `esp_ciphers` - (Optional) ESP cipher configuration for Phase 2 (data encryption). See [Cipher Config](#cipher-config) below.
- `name` - (Optional) The name of the connection.
- `tags` - (Optional) The list of tags to apply to the connection.
- `is_ipv6` - (Optional) Defines IP version of the IPSec Tunnel. Defaults to `false` (IPv4).
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the connection should be created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the connection is associated with.

### BGP Config

The `bgp_config_ipv4` and `bgp_config_ipv6` blocks support:

- `routing_policy_id` - (Required) The ID of the routing policy to use for BGP route filtering.
- `private_ip` - (Optional) The BGP peer IP on Scaleway side (within the IPSec tunnel), in CIDR notation (e.g., `169.254.0.1/30`). If not provided, Scaleway will assign it automatically.
- `peer_private_ip` - (Optional) The BGP peer IP on customer side (within the IPSec tunnel), in CIDR notation (e.g., `169.254.0.2/30`). If not provided, Scaleway will assign it automatically.

### Cipher Config

The `ikev2_ciphers` and `esp_ciphers` blocks support:

- `encryption` - (Required) The encryption algorithm.
- `integrity` - (Optional) The integrity/hash algorithm.
- `dh_group` - (Optional) The Diffie-Hellman group.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the connection.
- `status` - The status of the connection.
- `tunnel_status` - The status of the IPSec tunnel.
- `bgp_status_ipv4` - The status of the BGP IPv4 session.
- `bgp_status_ipv6` - The status of the BGP IPv6 session.
- `bgp_session_ipv4` - The BGP IPv4 session information. See [BGP Session](#bgp-session) below.
- `bgp_session_ipv6` - The BGP IPv6 session information. See [BGP Session](#bgp-session) below.
- `secret_id` - The ID of the secret containing the pre-shared key (PSK) for the connection.
- `secret_version` - The version of the secret containing the PSK.
- `route_propagation_enabled` - Whether route propagation is enabled.
- `created_at` - The date and time of the creation of the connection (RFC 3339 format).
- `updated_at` - The date and time of the last update of the connection (RFC 3339 format).
- `organization_id` - The Organization ID the connection is associated with.

### BGP Session

The `bgp_session_ipv4` and `bgp_session_ipv6` blocks contain (read-only):

- `routing_policy_id` - The routing policy ID used for this BGP session.
- `private_ip` - The BGP peer IP on Scaleway side (within the tunnel).
- `peer_private_ip` - The BGP peer IP on customer side (within the tunnel).

~> **Important:** Connections' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

~> **Important:** The pre-shared key (PSK) is auto-generated when the connection is created and stored in Scaleway Secret Manager. You can retrieve it using the `scaleway_secret_version` datasource or via the API.

## Retrieving the Pre-Shared Key (PSK)

The PSK is stored in Secret Manager and can be retrieved using:

```terraform
data "scaleway_secret_version" "s2s_psk" {
  secret_id = scaleway_s2s_vpn_connection.main.secret_id
  revision  = tostring(scaleway_s2s_vpn_connection.main.secret_version)
}

# The PSK is available as base64-encoded data
output "psk" {
  value     = data.scaleway_secret_version.s2s_psk.data
  sensitive = true
}
```

## Import

Connections can be imported using `{region}/{id}`, e.g.

```bash
terraform import scaleway_s2s_vpn_connection.main fr-par/11111111-1111-1111-1111-111111111111
```
