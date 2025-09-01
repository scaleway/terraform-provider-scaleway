---
subcategory: "Autoscaling"
page_title: "Scaleway: scaleway_autoscaling_instance_template"
---

# Resource: scaleway_autoscaling_instance_template

Books and manages Autoscaling Instance templates.

## Example Usage

### Basic

```terraform
resource "scaleway_autoscaling_instance_template" "main" {
  name            = "asg-template"
  commercial_type = "PLAY2-MICRO"
  tags            = ["terraform-test", "basic"]
  volumes {
    name        = "as-volume"
    volume_type = "sbs"
    boot        = true
    from_snapshot {
      snapshot_id = scaleway_block_snapshot.main.id
    }
    perf_iops = 5000
  }
  public_ips_v4_count = 1
  private_network_ids = [scaleway_vpc_private_network.main.id]
}
```

## Argument Reference

The following arguments are supported:

- `commercial_type` - (Required) The name of Instance commercial type.
- `tags` - (Optional) The tags associated with the Instance template.
- `name` - (Optional) The Instance group template.
- `image_id` - (Optional) The instance image ID. Can be an ID of a marketplace or personal image. This image must be compatible with `volume` and `commercial_type` template.
- `volumes` - (Required) The template of Instance volume.
    - `name` - The name of the volume.
    - `perf_iops` - The maximum IO/s expected, according to the different options available in stock (`5000 | 15000`).
    - `tags` - The list of tags assigned to the volume.
    - `boot` - Force the Instance to boot on this volume.
    - `volume_type` - The type of the volume.
- `security_group_id` - (Optional) The instance security group ID.
- `placement_group_id` - (Optional) The instance placement group ID. This is optional, but it is highly recommended to set a preference for Instance location within Availability Zone.
- `public_ips_v4_count` - (Optional) The number of flexible IPv4 addresses to attach to the new Instance.
- `public_ips_v6_count` - (Optional) The number of flexible IPv6 addresses to attach to the new Instance.
- `private_network_ids` - (Optional) The private Network IDs to attach to the new Instance.
- `cloud_init` - (Optional) The instance image ID. Can be an ID of a marketplace or personal image. This image must be compatible with `volume` and `commercial_type` template.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the Instance template exists.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the Project the Instance template is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Instance group.
- `created_at` - Date and time of Instance group's creation (RFC 3339 format).
- `updated_at` - Date and time of Instance group's last update (RFC 3339 format).

~> **Important:** Autoscaling Instance template IDs are [zonal](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

## Import

Autoscaling Instance templates can be imported using `{zone}/{id}`, e.g.

```bash
terraform import scaleway_autoscaling_instance_template.main fr-par-1/11111111-1111-1111-1111-111111111111
```
