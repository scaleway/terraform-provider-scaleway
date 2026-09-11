---
subcategory: "MessageQ"
page_title: "Scaleway: scaleway_messageq_deployment"
---

# Resource: scaleway_messageq_deployment

Creates and manages Scaleway MessageQ deployments.
For more information refer to the [API documentation](https://www.scaleway.com/en/developers/api/messageq).

~> **Note:** MessageQ is currently available as a `v1alpha1` API. Breaking changes may occur.

## Example Usage

### Basic

```terraform
resource "scaleway_messageq_deployment" "main" {
  name       = "my-messageq"
  version    = "4.0"
  node_count = 1
  node_type  = "MESSAGEQ-SHARED-2C-8G"
  user_name  = "admin"
  password   = "ThisIsASecurePassword123!"

  volume {
    type       = "sbs_5k"
    size_in_gb = 5
  }
}
```

### Production Setup with High Availability

```terraform
resource "scaleway_messageq_deployment" "prod" {
  name       = "orders-messageq"
  version    = "4.0"
  node_count = 3
  node_type  = "MESSAGEQ-DEDICATED-4C-16G"
  user_name  = "admin"
  password   = var.admin_password

  volume {
    type       = "sbs_15k"
    size_in_gb = 20
  }

  tags = ["env=prod", "team=platform"]
}
```

### With Private Network

```terraform
resource "scaleway_vpc" "main" {
  name = "my-vpc"
}

resource "scaleway_vpc_private_network" "pn" {
  name   = "my-private-network"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_messageq_deployment" "main" {
  name       = "my-messageq"
  version    = "4.0"
  node_count = 1
  node_type  = "MESSAGEQ-DEDICATED-2C-8G"
  user_name  = "admin"
  password   = "ThisIsASecurePassword123!"

  private_network {
    private_network_id = scaleway_vpc_private_network.pn.id
  }

  volume {
    type       = "sbs_5k"
    size_in_gb = 5
  }
}
```

## Argument Reference

The following arguments are supported:

- `version` - (Required, Forces new resource) MessageQ version to use. Changing this forces recreation of the deployment.
- `node_count` - (Required) Number of nodes in the cluster. Can be updated in-place via the Upgrade API.
- `node_type` - (Required, Forces new resource) Type of node to use. Changing this forces recreation of the deployment.
- `volume` - (Optional) Volume configuration for the cluster.
    - `type` - (Required, Forces new resource) Volume type. Valid values are `sbs_5k` (5K IOPS) or `sbs_15k` (15K IOPS). Changing this forces recreation of the deployment.
    - `size_in_gb` - (Required) Volume size in GB. Can be updated in-place via the Upgrade API.
- `password` - (Optional, Forces new resource, Sensitive) Bootstrap password for the MessageQ user. Only one of `password` or `password_wo` should be specified. Prefer [`scaleway_messageq_user`](messageq_user.md) for additional users and password rotation.
- `password_wo` - (Optional, Forces new resource, Write-only) Bootstrap password for the MessageQ user in [write-only](../guides/using-write-only-arguments.md) mode. Only one of `password` or `password_wo` should be specified, and `password_wo` is never stored in the Terraform state. Bootstrap credentials are immutable, so changing `password_wo_version` recreates the deployment: use [`scaleway_messageq_user`](messageq_user.md) to rotate a password in place.
- `password_wo_version` - (Optional, Forces new resource) Version of the write-only password. Required with `password_wo`.
- `name` - (Optional) Name of the MessageQ deployment. If not specified, a random name will be generated.
- `user_name` - (Optional, Forces new resource) Bootstrap username for the deployment. Prefer [`scaleway_messageq_user`](messageq_user.md) for additional users.
- `tags` - (Optional) List of tags to apply to the deployment.
- `private_network` - (Optional) Private network configuration. Can be added, updated, or removed on an existing deployment.
    - `private_network_id` - (Required) The ID of the private network. Format: `{region}/{id}` or just `{id}`.
- `region` - (Optional, Computed, Defaults to [provider](../index.md#arguments-reference) `region`) The [region](../guides/regions_and_zones.md#regions) in which the deployment should be created. Currently only `fr-par` is supported.
- `project_id` - (Defaults to [provider](../index.md#arguments-reference) `project_id`) The ID of the project the deployment is associated with.

~> **Note:** Without `private_network`, a public endpoint is created. Adding `private_network` to an existing deployment attaches a private endpoint but **does not remove the public one**, so the deployment stays reachable publicly. Only the endpoint matching your configuration is reported in `endpoints`; use the [MessageQ API](https://www.scaleway.com/en/developers/api/messageq) or the console to list every endpoint.

~> **Important:** Do not manage the same username both as bootstrap credentials on the deployment and as a `scaleway_messageq_user` resource.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the deployment in the format `{region}/{id}`.
- `status` - The status of the deployment (e.g., "ready", "creating", "upgrading").
- `created_at` - Date and time of deployment creation (RFC 3339 format).
- `updated_at` - Date and time of deployment last update (RFC 3339 format).
- `endpoints` - List of endpoints for accessing the deployment.
    - `id` - The ID of the endpoint.
    - `services` - List of services exposed on the endpoint.
        - `name` - Service name (e.g., AMQP, management).
        - `port` - Service port number.
        - `url` - Full URL to access the service.
    - `public` - Whether the endpoint is public (`true`) or private (`false`).
    - `private_network_id` - Private network ID if the endpoint is private.

## Import

MessageQ deployments can be imported using the `{region}/{id}`, e.g.

```bash
terraform import scaleway_messageq_deployment.main fr-par/11111111-1111-1111-1111-111111111111
```

~> **Note:** The API never returns `user_name` and `password`, so they cannot be recovered on import. If your configuration sets them, Terraform plans a replacement of the imported deployment because both force a new resource. Use `password_wo` (never stored in state, so never diffed) or `lifecycle { ignore_changes = [user_name, password] }` to import an existing deployment in place.
