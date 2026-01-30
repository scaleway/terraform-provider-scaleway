The `scaleway_s2s_vpn_connection_disable_route_propagation` action disables route propagation on an S2S VPN connection. This prevents any prefixes from being announced in the BGP session. Traffic will not be able to flow over the VPN gateway until route propagation is re-enabled (e.g. via the `scaleway_s2s_vpn_connection_enable_route_propagation` action).

Refer to the [S2S VPN documentation](https://www.scaleway.com/en/docs/network/s2s-vpn/) and [API documentation](https://www.scaleway.com/en/developers/api/s2s-vpn/) for more information.
