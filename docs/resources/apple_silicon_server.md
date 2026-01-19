---
subcategory: "Apple Silicon"
page_title: "Scaleway: scaleway_apple_silicon"
---

# Resource: scaleway_apple_silicon_server

The [`scaleway_apple_silicon_server`](https://registry.terraform.io/providers/scaleway/scaleway/latest/docs/resources/apple_silicon_server) resource creates and manages Scaleway Apple silicon servers.

For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/apple-silicon/).

## Example Usage

### Basic

```terraform
resource "scaleway_apple_silicon_server" "server" {
  name = "test-m1"
  type = "M1-M"
}
```

### Enable VPC and attach private network

```terraform
resource "scaleway_vpc" "vpc-apple-silicon" {
  name = "vpc-apple-silicon"
}

resource "scaleway_vpc_private_network" "pn-apple-silicon" {
  name = "pn-apple-silicon"
  vpc_id = scaleway_vpc.vpc-apple-silicon.id
}

resource "scaleway_apple_silicon_server" "my-server" {
    name = "TestAccServerEnableVPC"
    type = "M2-M"
    enable_vpc = true
    private_network {
      id = scaleway_vpc_private_network.pn-apple-silicon.id
    }
}
```

### With `github` runner

```terraform
data "scaleway_apple_silicon_os" "by_name" {
  name = "devos-sequoia-15.6"
}

resource "scaleway_apple_silicon_runner" "main" {
    name       = "TestAccRunnerGithub"
    ci_provider   = "github"
    url        = "https://github.com/my-repo-url"
    token      = "MY_GITHUB_RUNNER_TOKEN"
}

resource scaleway_apple_silicon_server main {
    name = "TestAccServerRunner"
    type = "M2-L"
    public_bandwidth = 1000000000
    os_id = data.scaleway_apple_silicon_os.by_name.id
    runner_ids = [scaleway_apple_silicon_runner.main.id]
}
```

## Argument Reference

The following arguments are supported:

- `type` - (Required) The commercial type of the server. You find all the available types on
  the [pricing page](https://www.scaleway.com/en/pricing/apple-silicon/). Updates to this field will recreate a new
  resource.

- `name` - (Optional) The name of the server.

- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which
  the server should be created.

- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the server is
  associated with.

- `enable_vpc` - (Optional, Default: false): Enables the VPC option when set to true.

- `os_id` - (Optional) The ID of the OS to use for the server.

- `runner_ids` - (Optional) List of runner IDs to assign to the server. At the moment, only a single runner can be attached to a server. Compatible only with runners of type `github` and `gitlab`, with the `devos-sequoia-15.6` offer and `M2-L` server type

- `private_network` - (Optional) The private networks to attach to the server
    - `id` - The private network ID
    - `ipam_ip_ids` - A list of IPAM IP IDs to attach to the server.

- `commitment_type` (Optional, Default: duration_24h): Activate commitment for this server

- `public_bandwidth` (Optional) Configure the available public bandwidth for your server in bits per second. This option may not be available for all offers.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the server.

~> **Important:** Apple Silicon servers' IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

- `state` - The state of the server.
- `ip` - IPv4 address of the server (IPv4 address).
- `vnc_url` - URL of the VNC.
- `private_ips` - The list of private IPv4 and IPv6 addresses associated with the server.
    - `id` - The ID of the IP address resource.
    - `address` - The private IP address.
- `created_at` - The date and time of the creation of the Apple Silicon server.
- `updated_at` - The date and time of the last update of the Apple Silicon server.
- `deleted_at` - The minimal date and time on which you can delete this server due to Apple licence.
- `organization_id` - The organization ID the server is associated with.
- `vpc_status` - The current status of the VPC option.
- `private_network` - The private networks to attach to the server
    - `vlan`  - The VLAN ID associated with the private network.
    - `status` - The current status of the private network.
    - `created_at` - The date and time the private network was created.
    - `updated_at` - The date and time the private network was last updated.

## Import

Instance servers can be imported using the `{zone}/{id}`, e.g.

```bash
terraform import scaleway_apple_silicon_server.main fr-par-1/11111111-1111-1111-1111-111111111111
```
