---
layout: "scaleway"
page_title: "Scaleway: scaleway_lb_backend_beta"
description: |-
  Manages Scaleway Load-Balancer Backends.
---

# scaleway_lb_backend_beta

-> **Note:** This terraform resource is flagged beta and might include breaking change in future releases.

Creates and manages Scaleway Load-Balancer Backends. For more information, see [the documentation](https://developers.scaleway.com/en/products/lb/api).

## Examples
    
### Basic

```hcl
resource "scaleway_lb_backend_beta" "backend01" {
    lb_id = scaleway_lb_beta.lb01.id
    name = "backend01"
    forward_protocol = "http"
    forward_port = "80"
}
```

## Arguments Reference

The following arguments are supported:

- `lb_id`                       - (Required) The load-balancer ID this backend is attached to.
~> **Important:** Updates to `lb_id` will recreate the backend.
- `forward_protocol`            - (Required) Backend protocol. Possible values are: `TCP` or `HTTP`.
- `name`                        - (Optional) The name of the load-balancer backend.
- `forward_port`                - (Required) User sessions will be forwarded to this port of backend servers.
- `forward_port_algorithm`      - (Default: `roundrobin`) Load balancing algorithm. Possible values are: `roundrobin` and `leastconn`.
- `sticky_sessions`             - (Default: `none`) Load balancing algorithm. Possible values are: `none`, `cookie` and `table`.
- `sticky_sessions_cookie_name` - (Optional) Cookie name for for sticky sessions. Only applicable when sticky_sessions is set to `cookie`.
- `server_ips`                  - (Optional) List of backend server IP addresses. Addresses can be either IPv4 or IPv6.
- `send_proxy_v2`               - (Default: `false`) Enables PROXY protocol version 2.
- `timeout_server`              - (Optional) Maximum server connection inactivity time. (e.g.: `1s`)
- `timeout_connect`             - (Optional) Maximum initial server connection establishment time. (e.g.: `1s`)
- `timeout_tunnel`              - (Optional) Maximum tunnel inactivity time. (e.g.: `1s`)
- `on_marked_down_action`       - (Default: `none`) Modify what occurs when a backend server is marked down. Possible values are: `none` and `shutdown_sessions`.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the loadbalancer backend.


## Import

Load-Balancer backend can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_lb_backend_beta.backend01 fr-par/11111111-1111-1111-1111-111111111111
```
