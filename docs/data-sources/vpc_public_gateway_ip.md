---
page_title: "Scaleway: scaleway_vpc_public_gateway_ip"
description: |-
  Get information about Scaleway VPC Public Gateway IPs.
---

# scaleway_vpc_public_gateway_ip

Gets information about a public gateway IP.

## Example Usage

```hcl
resource "scaleway_vpc_public_gateway_ip" "main" {
}

data "scaleway_vpc_public_gateway_ip" "ip_by_id" {
    ip_id = "${scaleway_vpc_public_gateway_ip.main.id}"
}
```

## Argument Reference

## Attributes Reference

`id` is set to the ID of the found public gateway ip. Addition attributes are
exported.
