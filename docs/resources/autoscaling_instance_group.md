---
subcategory: "Autoscaling"
page_title: "Scaleway: scaleway_autoscaling_instance_group"
---

# Resource: scaleway_autoscaling_instance_group

Books and manages Autoscaling Instance groups.

## Example Usage

### Basic

```terraform
resource "scaleway_autoscaling_instance_group" "main" {
  name        = "asg-group"
  template_id = scaleway_autoscaling_instance_template.main.id
  tags        = ["terraform-test", "instance-group"]
  capacity {
    max_replicas   = 5
    min_replicas   = 1
    cooldown_delay = "300"
  }
  load_balancer {
    id                 = scaleway_lb.main.id
    backend_ids        = [scaleway_lb_backend.main.id]
    private_network_id = scaleway_vpc_private_network.main.id
  }
}
```

### With template and policies

```terraform
resource "scaleway_vpc" "main" {
  name = "TestAccAutoscalingVPC"
}

resource "scaleway_vpc_private_network" "main" {
  name   = "TestAccAutoscalingVPC"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_block_volume" "main" {
  iops       = 5000
  size_in_gb = 10
}

resource "scaleway_block_snapshot" "main" {
  name      = "test-ds-block-snapshot-basic"
  volume_id = scaleway_block_volume.main.id
}

resource "scaleway_lb_ip" "main" {}
resource "scaleway_lb" "main" {
  ip_id = scaleway_lb_ip.main.id
  name  = "test-lb"
  type  = "lb-s"
  private_network {
    private_network_id = scaleway_vpc_private_network.main.id
  }
}

resource "scaleway_lb_backend" "main" {
  lb_id            = scaleway_lb.main.id
  forward_protocol = "tcp"
  forward_port     = 80
  proxy_protocol   = "none"
}

resource "scaleway_autoscaling_instance_template" "main" {
  name            = "autoscaling-instance-template-basic"
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

resource "scaleway_autoscaling_instance_group" "main" {
  name        = "autoscaling-instance-group-basic"
  template_id = scaleway_autoscaling_instance_template.main.id
  tags        = ["terraform-test", "instance-group"]
  capacity {
    max_replicas   = 5
    min_replicas   = 1
    cooldown_delay = "300"
  }
  load_balancer {
    id                 = scaleway_lb.main.id
    backend_ids        = [scaleway_lb_backend.main.id]
    private_network_id = scaleway_vpc_private_network.main.id
  }
  delete_servers_on_destroy = true
}

resource "scaleway_autoscaling_instance_policy" "up" {
  instance_group_id = scaleway_autoscaling_instance_group.main.id
  name              = "scale-up-if-cpu-high"
  action            = "scale_up"
  type              = "flat_count"
  value             = 1
  priority          = 1

  metric {
    name               = "cpu scale up"
    managed_metric     = "managed_metric_instance_cpu"
    operator           = "operator_greater_than"
    aggregate          = "aggregate_average"
    sampling_range_min = 5
    threshold          = 70
  }
}

resource "scaleway_autoscaling_instance_policy" "down" {
  instance_group_id = scaleway_autoscaling_instance_group.main.id
  name              = "scale-down-if-cpu-low"
  action            = "scale_down"
  type              = "flat_count"
  value             = 1
  priority          = 2

  metric {
    name               = "cpu scale down"
    managed_metric     = "managed_metric_instance_cpu"
    operator           = "operator_less_than"
    aggregate          = "aggregate_average"
    sampling_range_min = 5
    threshold          = 40
  }
}

```

## Argument Reference

The following arguments are supported:

- `template_id` - (Required) The ID of the Instance template to attach to the Instance group.
- `tags` - (Optional) The tags associated with the Instance group.
- `name` - (Optional) The Instance group name.
- `capacity` - (Optional) The specification of the minimum and maximum replicas for the Instance group, and the cooldown interval between two scaling events.
    - `max_replicas` - The maximum count of Instances for the Instance group.
    - `min_replicas` - The minimum count of Instances for the Instance group.
    - `cooldown_delay` - Time (in seconds) after a scaling action during which requests to carry out a new scaling action will be denied.
- `load_balancer` - (Optional) The specification of the Load Balancer to link to the Instance group.
    - `id` - The ID of the Load Balancer.
    - `backend_ids` - The Load Balancer backend IDs.
    - `private_network_id` - The ID of the Private Network attached to the Load Balancer.
- `delete_servers_on_destroy` - (Optional) Whether to delete all instances in this group when the group is destroyed. Set to `true` to tear them down, `false` (the default) leaves them running.
- `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the Instance group exists.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the Project the Instance group is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the Instance group.
- `created_at` - Date and time of Instance group's creation (RFC 3339 format).
- `updated_at` - Date and time of Instance group's last update (RFC 3339 format).

~> **Important:** Autoscaling Instance group IDs are [zonal](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

## Import

Autoscaling Instance groups can be imported using `{zone}/{id}`, e.g.

```bash
terraform import scaleway_autoscaling_instance_group.main fr-par-1/11111111-1111-1111-1111-111111111111
```
