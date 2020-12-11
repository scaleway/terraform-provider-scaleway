---
page_title: "Scaleway: scaleway_lb_backend"
description: |-
  Manages Scaleway Load-Balancer Backends.
---

# scaleway_lb_backend

-> **Note:** This terraform resource is flagged beta and might include breaking change in future releases.

Creates and manages Scaleway Load-Balancer Backends. For more information, see [the documentation](https://developers.scaleway.com/en/products/lb/api).

## Examples

### Basic

```hcl
resource "scaleway_lb_backend" "backend01" {
  lb_id            = scaleway_lb.lb01.id
  name             = "backend01"
  forward_protocol = "http"
  forward_port     = "80"
}
```

### With HTTP Health Check

```hcl
resource "scaleway_lb_backend" "backend01" {
  lb_id            = scaleway_lb.lb01.id
  name             = "backend01"
  forward_protocol = "http"
  forward_port     = "80"

  health_check_http {
    uri = "www.test.com/health"
  }
}
```

## Arguments Reference

The following arguments are supported:

### Basic arguments

- `lb_id`                       - (Required) The load-balancer ID this backend is attached to.
~> **Important:** Updates to `lb_id` will recreate the backend.
- `forward_protocol`            - (Required) Backend protocol. Possible values are: `tcp` or `http`.
- `name`                        - (Optional) The name of the load-balancer backend.
- `forward_port`                - (Required) User sessions will be forwarded to this port of backend servers.
- `forward_port_algorithm`      - (Default: `roundrobin`) Load balancing algorithm. Possible values are: `roundrobin`, `leastconn` and `first`.
- `sticky_sessions`             - (Default: `none`) Load balancing algorithm. Possible values are: `none`, `cookie` and `table`.
- `sticky_sessions_cookie_name` - (Optional) Cookie name for for sticky sessions. Only applicable when sticky_sessions is set to `cookie`.
- `server_ips`                  - (Optional) List of backend server IP addresses. Addresses can be either IPv4 or IPv6.
- `send_proxy_v2`               - DEPRECATED please use `proxy_protocol` instead - (Default: `false`) Enables PROXY protocol version 2.
- `proxy_protocol`              - (Default: `none`) Choose the type of PROXY protocol to enable (`none`, `v1`, `v2`, `v2_ssl`, `v2_ssl_cn`)
- `timeout_server`              - (Optional) Maximum server connection inactivity time. (e.g.: `1s`)
- `timeout_connect`             - (Optional) Maximum initial server connection establishment time. (e.g.: `1s`)
- `timeout_tunnel`              - (Optional) Maximum tunnel inactivity time. (e.g.: `1s`)

### Health Check arguments

Backends use Health Check to test if a backend server is ready to receive requests.
You may use one of the following health check types: `TCP`, `HTTP` or `HTTPS`. (Default: `TCP`)

- `health_check_timeout`        - (Default: `30s`) Timeout before we consider a HC request failed.
- `health_check_delay`          - (Default: `60s`) Interval between two HC requests.
- `health_check_port`           - (Default: `forward_port`) Port the HC requests will be send to.
- `health_check_max_retries`    - (Default: `2`) Number of allowed failed HC requests before the backend server is marked down.
- `health_check_tcp`            - (Optional) This block enable TCP health check. Only one of `health_check_tcp`, `health_check_http` and `health_check_https` should be specified.
- `health_check_http`           - (Optional) This block enable HTTP health check. Only one of `health_check_tcp`, `health_check_http` and `health_check_https` should be specified.
    - `uri`                       - (Required) The HTTP endpoint URL to call for HC requests.
    - `method`                    - (Default: `GET`) The HTTP method to use for HC requests.
    - `code`                      - (Default: `200`) The expected HTTP status code.
- `health_check_https`          - (Optional) This block enable HTTPS health check. Only one of `health_check_tcp`, `health_check_http` and `health_check_https` should be specified.
    - `uri`                       - (Required) The HTTPS endpoint URL to call for HC requests.
    - `method`                    - (Default: `GET`) The HTTP method to use for HC requests.
    - `code`                      - (Default: `200`) The expected HTTP status code.
- `on_marked_down_action`       - (Default: `none`) Modify what occurs when a backend server is marked down. Possible values are: `none` and `shutdown_sessions`.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the loadbalancer backend.


## Import

Load-Balancer backend can be imported using the `{region}/{id}`, e.g.

```bash
$ terraform import scaleway_lb_backend.backend01 fr-par/11111111-1111-1111-1111-111111111111
```
