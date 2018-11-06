---
layout: "scaleway"
page_title: "Scaleway: ip_reverse_dns"
sidebar_current: "docs-scaleway-resource-ip-reverse-dns"
description: |-
  Manages Scaleway IPs.
---

# scaleway_ip_reverse_dns

Provides reverse DNS settings for IPs.
For additional details please refer to [API documentation](https://developer.scaleway.com/#ips).

## Example Usage

```hcl
resource "scaleway_ip" "test_service" {}

resource "scaleway_ip_reverse_dns" "google" {
  ip = "${scaleway_ip.test_service.id}"
  reverse = "test_service.awesome-corp.com"
}
```

## Argument Reference

The following arguments are supported:

* `ip` - (Required) ID or Address of IP 
* `reverse` - (Required) Reverse DNS of the IP

## Attributes Reference

The following attributes are exported:

* `id` - ID of the new resource
* `reverse` - reverse DNS setting of the IP resource
