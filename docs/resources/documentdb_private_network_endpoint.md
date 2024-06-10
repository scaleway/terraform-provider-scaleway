---
subcategory: "Databases"
page_title: "Scaleway: scaleway_documentdb_private_network_endpoint"
---

# Resource: scaleway_documentdb_private_network_endpoint

Creates and manages Scaleway Database Private Network Endpoint.

## Example Usage

### Example Basic

```terraform
resource "scaleway_vpc_private_network" "pn" {
  name = "my_private_network"
}

resource "scaleway_documentdb_private_network_endpoint" "main" {
  instance_id    = "11111111-1111-1111-1111-111111111111"
  private_network {
    ip_net = "172.16.32.3/22"
    id     = scaleway_vpc_private_network.pn.id
  }
  depends_on = [scaleway_vpc_private_network.pn]
}
```

## Argument Reference

The following arguments are supported:

- `instance_id` - (Required) UUID of the documentdb instance.

- `private_network` - (Optional) The private network specs details. This is a list with maximum one element and supports the following attributes:
    - `id` - (Required) The private network ID.
    - `ip_net` - (Optional) The IP network address within the private subnet. This must be an IPv4 address with a CIDR notation. The IP network address within the private subnet is determined by the IP Address Management (IPAM) service if not set.
    - `ip` - (Computed) The IP of your private network service.
    - `port` - (Optional, Computed) The port of your private service.
    - `name` - (Computed) The name of your private service.
    - `hostname` - (Computed) The hostname of your endpoint.
    - `zone` - (Computed) The zone of your endpoint.

- `region` - (Optional) The region of the endpoint.


~> **NOTE:** Please calculate your host IP.
using [cirhost](https://developer.hashicorp.com/terraform/language/functions/cidrhost). Otherwise, lets IPAM service
handle the host IP on the network.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `ip` - IPv4 address on the network.
- `port` - Port in the Private Network.
- `name` - Name of the endpoint.
- `hostname` - Hostname of the endpoint.


~> **Important:** Database instances' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they
are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

## Import

Database Instance Endpoint can be imported using the `{region}/{endpoint_id}`, e.g.

```bash
$ terraform import scaleway_documentdb_private_network_endpoint.end fr-par/11111111-1111-1111-1111-111111111111
```
