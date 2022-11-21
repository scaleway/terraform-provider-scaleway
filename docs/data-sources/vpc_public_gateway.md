---
page_title: "Scaleway: scaleway_vpc_public_gateway"
description: |-
Get information about Scaleway VPC Public Gateways.
---

# scaleway_vpc_public_gateway

Gets information about a public gateway.

## Example Usage

```hcl
resource "scaleway_vpc_public_gateway" "main" {
  name = "demo"
  type = "VPC-GW-S"
  zone = "nl-ams-1"
}

data "scaleway_vpc_public_gateway" "pg_test_by_name" {
  name = "${scaleway_vpc_public_gateway.main.name}"
  zone = "nl-ams-1"
}

data "scaleway_vpc_public_gateway" "pg_test_by_id" {
  public_gateway_id = "${scaleway_vpc_public_gateway.main.id}"
}
```

## Argument Reference

- `name` - (Required) Exact name of the public gateway.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which
  the public gateway should be created.

## Attributes Reference

`id` is set to the ID of the found public gateway. Addition attributes are
exported.
