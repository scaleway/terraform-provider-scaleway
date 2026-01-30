The `scaleway_s2s_vpn_connection_enable_route_propagation` action enables route propagation on an S2S VPN connection. This allows all allowed prefixes (defined in a routing policy) to be announced in the BGP session, so that traffic can flow between the attached VPC and the on-premises infrastructure along the announced routes.

Note that by default, even when route propagation is enabled, all routes are blocked. It is essential to attach a routing policy to the connection to define the ranges of routes to announce.

Refer to the [S2S VPN documentation](https://www.scaleway.com/en/docs/network/s2s-vpn/) and [API documentation](https://www.scaleway.com/en/developers/api/s2s-vpn/) for more information.
