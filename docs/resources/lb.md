---
page_title: "Scaleway: scaleway_lb"
description: |-
  Manages Scaleway Load-Balancers.
---

# scaleway_lb

Creates and manages Scaleway Load-Balancers.
For more information, see [the documentation](https://developers.scaleway.com/en/products/lb/zoned_api).

## Examples

### Basic

```hcl
resource "scaleway_lb_ip" "ip" {
}

resource "scaleway_lb" "base" {
  ip_id  = scaleway_lb_ip.ip.id
  zone = "fr-par-1"
  type   = "LB-S"
}
```

## Arguments Reference

The following arguments are supported:

- `ip_id` - (Required) The ID of the associated IP. See below.

~> **Important:** Updates to `ip_id` will recreate the load-balancer.

- `type` - (Required) The type of the load-balancer.

~> **Important:** Updates to `type` will recreate the load-balancer.

- `name` - (Optional) The name of the load-balancer.

- `tags` - (Optional) The tags associated with the load-balancers.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the IP should be reserved.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the load-balancer is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the load-balancer.
- `ip_address` -  The load-balance public IP Address
- `organization_id` - The organization ID the load-balancer is associated with.

## IP ID

Since v1.15.0, `ip_id` is a required field. This means that now a separate `scaleway_lb_ip` is required.
When importing, the IP needs to be imported as well as the LB.
When upgrading to v1.15.0, you will need to create a new `scaleway_lb_ip` resource and import it.

For instance, if you had the following:

```hcl
resource "scaleway_lb" "base" {
  zone = "fr-par-1"
  type   = "LB-S"
}
```

You will need to update it to:

```hcl
resource "scaleway_lb_ip" "ip" {
}

resource "scaleway_lb" "base" {
  ip_id  = scaleway_lb_ip.ip.id
  zone = "fr-par-1"
  type   = "LB-S"
}
```

And before running `terraform apply` you will need to import the IP with:

```bash
$ terraform import scaleway_lb_ip.ip fr-par/11111111-1111-1111-1111-111111111111
```

The IP ID can either be found in the console, or you can run:

```bash
$ terraform state show scaleway_lb.base
```

and look for `ip_id`.

## Import

Load-Balancer can be imported using the `{zone}/{id}`, e.g.

```bash
$ terraform import scaleway_lb.lb01 fr-par-1/11111111-1111-1111-1111-111111111111
```

Be aware that you will also need to import the `scaleway_lb_ip` resource.
