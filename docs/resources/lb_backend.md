---
page_title: "scaleway_lb_backend Resource - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Resource `scaleway_lb_backend`





## Schema

### Required

- **forward_port** (Number) User sessions will be forwarded to this port of backend servers
- **forward_protocol** (String) Backend protocol
- **lb_id** (String) The load-balancer ID

### Optional

- **forward_port_algorithm** (String) Load balancing algorithm
- **health_check_delay** (String) Interval between two HC requests
- **health_check_http** (Block List, Max: 1) (see [below for nested schema](#nestedblock--health_check_http))
- **health_check_https** (Block List, Max: 1) (see [below for nested schema](#nestedblock--health_check_https))
- **health_check_max_retries** (Number) Number of allowed failed HC requests before the backend server is marked down
- **health_check_port** (Number) Port the HC requests will be send to. Default to `forward_port`
- **health_check_tcp** (List of String)
- **health_check_timeout** (String) Timeout before we consider a HC request failed
- **id** (String) The ID of this resource.
- **name** (String) The name of the backend
- **on_marked_down_action** (String) Modify what occurs when a backend server is marked down
- **proxy_protocol** (String) Type of PROXY protocol to enable
- **send_proxy_v2** (Boolean, Deprecated) Enables PROXY protocol version 2
- **server_ips** (List of String) Backend server IP addresses list (IPv4 or IPv6)
- **sticky_sessions** (String) Load balancing algorithm
- **sticky_sessions_cookie_name** (String) Cookie name for for sticky sessions
- **timeout_connect** (String) Maximum initial server connection establishment time
- **timeout_server** (String) Maximum server connection inactivity time
- **timeout_tunnel** (String) Maximum tunnel inactivity time
- **timeouts** (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

<a id="nestedblock--health_check_http"></a>
### Nested Schema for `health_check_http`

Required:

- **uri** (String) The HTTP endpoint URL to call for HC requests

Optional:

- **code** (Number) The expected HTTP status code
- **method** (String) The HTTP method to use for HC requests


<a id="nestedblock--health_check_https"></a>
### Nested Schema for `health_check_https`

Required:

- **uri** (String) The HTTPS endpoint URL to call for HC requests

Optional:

- **code** (Number) The expected HTTP status code
- **method** (String) The HTTP method to use for HC requests


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- **default** (String)


