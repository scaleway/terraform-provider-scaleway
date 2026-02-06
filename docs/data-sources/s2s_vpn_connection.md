---
subcategory: "S2S VPN"
page_title: "Scaleway: scaleway_s2s_vpn_connection"
---

# scaleway_s2s_vpn_connection

Gets information about a Site-to-Site VPN Connection.

For further information refer to the Site-to-Site VPN [API documentation](https://www.scaleway.com/en/developers/api/site-to-site-vpn/).

## Example Usage

```hcl
# Get info by name
data "scaleway_s2s_vpn_connection" "my_connection" {
  name = "foobar"
}

# Get info by connection ID
data "scaleway_s2s_vpn_connection" "my_connection" {
  connection_id = "11111111-1111-1111-1111-111111111111"
}
```

## Argument Reference

- `name` - (Optional) The name of the connection.

- `connection_id` - (Optional) The connection ID.

  -> **Note** You must specify at least one: `name` and/or `connection_id`.

- `region` - (Defaults to [provider](../index.md) `region`) The [region](../guides/regions_and_zones.md#regions) in which the connection exists.

- `project_id` - (Optional) The ID of the project the connection is associated with.

## Attributes Reference

In addition to all above arguments, the following attributes are exported:

- `id` - The ID of the connection.
- `vpn_gateway_id` - The ID of the VPN gateway attached to the connection.
- `customer_gateway_id` - The ID of the customer gateway attached to the connection.
- `is_ipv6` - Whether the IPSec tunnel uses IPv6.
- `initiation_policy` - Who initiates the IPSec tunnel.
- `enable_route_propagation` - Whether route propagation is enabled.
- `route_propagation_enabled` - Whether route propagation is currently enabled.
- `ikev2_ciphers` - The IKEv2 ciphers configuration.
- `esp_ciphers` - The ESP ciphers configuration.
- `bgp_config_ipv4` - The BGP IPv4 configuration.
- `bgp_config_ipv6` - The BGP IPv6 configuration.
- `bgp_session_ipv4` - The BGP IPv4 session information.
- `bgp_session_ipv6` - The BGP IPv6 session information.
- `bgp_status_ipv4` - The status of the BGP IPv4 session.
- `bgp_status_ipv6` - The status of the BGP IPv6 session.
- `status` - The status of the connection.
- `tunnel_status` - The status of the IPSec tunnel.
- `secret_id` - The ID of the secret containing the pre-shared key (PSK).
- `secret_version` - The version of the secret containing the PSK.
- `tags` - The tags associated with the connection.
- `created_at` - The date and time of creation of the connection.
- `updated_at` - The date and time of the last update of the connection.
- `organization_id` - The Organization ID the connection is associated with.

~> **Important:** Connections IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`
