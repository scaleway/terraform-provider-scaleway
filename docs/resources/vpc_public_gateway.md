---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_public_gateway"
---

# Resource: scaleway_vpc_public_gateway

Creates and manages Scaleway Public Gateways.
For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/public-gateway).

## Example Usage

### Basic

```terraform
resource "scaleway_vpc_public_gateway" "main" {
    name = "public_gateway_demo"
    type = "VPC-GW-S"
    tags = ["demo", "terraform"]
}
```

### With bastion

```terraform
resource "scaleway_iam_ssh_key" "key1" {
  name       = "key1"
  public_key = file("~/.ssh/id_rsa.pub")
}

resource "scaleway_iam_ssh_key" "key2" {
  name       = "key2"
  public_key = file("~/.ssh/another_key.pub")}

# Use a local variable to compute a hash of the SSH keys
locals {
  ssh_keys_hash = sha256(join(",", [
    scaleway_iam_ssh_key.key1.public_key,
    scaleway_iam_ssh_key.key2.public_key,
  ]))
}

resource "scaleway_vpc_public_gateway" "main" {
    name             = "public_gateway_demo"
    type             = "VPC-GW-S"
    tags             = ["demo", "terraform"]
    bastion_enabled  = true
    bastion_port     = 61000
    refresh_ssh_keys = local.ssh_keys_hash
}
```

## Argument Reference

The following arguments are supported:

- `type` - (Required) The gateway type.
- `name` - (Optional) The name for the Public Gateway. If not provided it will be randomly generated.
- `tags` - (Optional) The tags to associate with the Public Gateway.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the Public Gateway should be created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the public gateway is associated with.
- `upstream_dns_servers` - (Optional) Override the gateway's default recursive DNS servers, if DNS features are enabled.
- `ip_id` - (Optional) Attach an existing flexible IP to the gateway.
- `bastion_enabled` - (Optional) Enable SSH bastion on the gateway.
- `bastion_port` - (Optional) The port on which the SSH bastion will listen.
- `enable_smtp` - (Optional) Enable SMTP on the gateway.
- `refresh_ssh_keys` - (Optional) Trigger a refresh of the SSH keys on the Public Gateway by changing this field's value.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Public Gateway.

~> **Important:** Public Gateways' IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

- `organization_id` - The Organization ID the Public Gateway is associated with.
- `created_at` - The date and time of the creation of the Public Gateway.
- `updated_at` - The date and time of the last update of the Public Gateway.
- `status` - The status of the public gateway.

## Import

Public Gateways can be imported using `{zone}/{id}`, e.g.

```bash
terraform import scaleway_vpc_public_gateway.main fr-par-1/11111111-1111-1111-1111-111111111111
```
