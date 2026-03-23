---
subcategory: "VPC"
page_title: "Scaleway: scaleway_vpc_connector"
---

# Resource: scaleway_vpc_connector

Creates and manages Scaleway VPC Connectors.
For more information, see [the main documentation](https://www.scaleway.com/en/docs/vpc/concepts/).

## Example Usage

### Basic

```terraform
resource "scaleway_vpc" "vpc01" {
  name = "my-vpc-source"
}

resource "scaleway_vpc" "vpc02" {
  name = "my-vpc-target"
}

resource "scaleway_vpc_connector" "main" {
  name          = "my-vpc-connector"
  vpc_id        = scaleway_vpc.vpc01.id
  target_vpc_id = scaleway_vpc.vpc02.id
}
```

### With Tags

```terraform
resource "scaleway_vpc" "vpc01" {
  name = "my-vpc-source"
}

resource "scaleway_vpc" "vpc02" {
  name = "my-vpc-target"
}

resource "scaleway_vpc_connector" "main" {
  name          = "my-vpc-connector"
  vpc_id        = scaleway_vpc.vpc01.id
  target_vpc_id = scaleway_vpc.vpc02.id
  tags          = ["production", "connector"]
}
```

## Argument Reference

The following arguments are supported:

- `vpc_id` - (Required) The ID of the source VPC.
- `target_vpc_id` - (Required) The ID of the target VPC to connect to.
- `name` - (Optional) The name of the VPC connector. If not provided it will be randomly generated.
- `tags` - (Optional) The tags to associate with the VPC connector.
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) of the VPC connector.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the Project the VPC connector is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the VPC connector.
- `status` - The status of the VPC connector.
- `created_at` - The date and time of the creation of the VPC connector (RFC 3339 format).
- `updated_at` - The date and time of the last update of the VPC connector (RFC 3339 format).

~> **Important:** VPC connectors' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111

- `organization_id` - The Organization ID the VPC connector is associated with.

## Import

VPC connectors can be imported using `{region}/{id}`, e.g.

```bash
terraform import scaleway_vpc_connector.main fr-par/11111111-1111-1111-1111-111111111111
```
