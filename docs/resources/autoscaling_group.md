---
subcategory: "Autoscaling"
page_title: "Scaleway: scaleway_autoscaling_group"
---

# Resource: scaleway_autoscaling_group

Manages Autoscaling Groups.

## Example Usage

### CPU Target

```terraform
resource "scaleway_instance_template" "main" {
  server_type = "GP1-L"
}

resource "scaleway_autoscaling_group" "main" {
  template_id = scaleway_instance_template.main.id
  tags        = ["tf-examples", "scaleway_autoscaling_group", "cpu-target"]
  name        = "asg-cpu-target"

  scaling_policy = {
    minimum_size       = 0
    maximum_size       = 25
    scale_in_cooldown  = "2m"
    scale_out_cooldown = "1m"
    scale_in_step      = 5
    scale_out_step     = 5

    cpu_target = 92
  }
}
```

### Fixed size

```terraform
resource "scaleway_instance_template" "main" {
  server_type = "PRO2-M"
}

resource "scaleway_autoscaling_group" "main" {
  template_id = scaleway_instance_template.main.id
  tags        = ["tf-examples", "scaleway_autoscaling_group", "fixed-size"]
  name        = "asg-fixed-size"

  scaling_policy = {
    minimum_size       = 1
    maximum_size       = 20
    scale_in_cooldown  = "7m"
    scale_out_cooldown = "30s"

    fixed_size = 10
  }
}
```

### Load Balancer Configuration

```terraform
# VPC resources
resource "scaleway_vpc" "vpc" {}
resource "scaleway_vpc_private_network" "pn" {
  vpc_id = scaleway_vpc.vpc.id
}

# LB resources
resource "scaleway_lb" "lb" {
  type = "LB-S"
}
resource "scaleway_lb_backend" "back0" {
  lb_id            = scaleway_lb.lb.id
  forward_protocol = "http"
  forward_port     = 80
}
resource "scaleway_lb_backend" "back1" {
  lb_id            = scaleway_lb.lb.id
  forward_protocol = "tcp"
  forward_port     = 81
}

# Template
resource "scaleway_instance_template" "main" {
  tags        = ["terraform-test", "scaleway_autoscaling_group", "lb_config"]
  server_type = "PRO2-M"
}

# AutoScaling Group
resource "scaleway_autoscaling_group" "main" {
  template_id = scaleway_instance_template.main.id

  scaling_policy = {
    minimum_size  = 0
    maximum_size  = 25
    memory_target = 75
  }

  load_balancer_configuration = {
    load_balancer_id = scaleway_lb.lb.id
    backends = [{
      backend_id     = scaleway_lb_backend.back0.id
      address_family = "ipv6"
      }, {
      backend_id         = scaleway_lb_backend.back1.id
      address_family     = "ipv4"
      private_network_id = scaleway_vpc_private_network.pn.id
    }]
    auto_healing = {
      enabled      = true
      grace_period = "10m"
    }
  }
}
```


## Argument Reference

The following arguments are supported:

- `template_id` - (Required) The ID of the Instance Template used to create instances in this group.

- `name` - (Optional) The name of the AutoScaling Group. If not set, a name will be generated.

- `tags` - (Optional) The tags associated with the AutoScaling Group.

- `scaling_policy` - (Required) The scaling policy configuration.

  -> The `scaling_policy` block contains:

    - `minimum_size` - (Required) The minimum number of instances in the group.

    - `maximum_size` - (Required) The maximum number of instances in the group.

    - `scale_in_cooldown` - (Optional, Defaults to `1m`) The cooldown duration after a scale-in event.

    - `scale_out_cooldown` - (Optional, Defaults to `1m`) The cooldown duration after a scale-out event.

    - `scale_in_step` - (Optional, Defaults to `1`) The number of instances to remove during scale-in event.

    - `scale_out_step` - (Optional, Defaults to `1`) The number of instances to add during scale-out event.

    - `fixed_size` - (Optional, Conflicts with `cpu_target` and `memory_target`) The fixed number of instances for the group.

    - `cpu_target` - (Optional, Conflicts with `fixed_size` and `memory_target`) The target CPU utilization percentage to trigger scaling events.

    - `memory_target` - (Optional, Conflicts with `fixed_size` and `cpu_target`) The target memory utilization percentage to trigger scaling events.

  ~> **Important:** Exactly one of `fixed_size`, `cpu_target` and `memory_target` must be defined.

- `load_balancer_configuration` - (Optional) The load balancer configuration.

  ~> **Important:** Due to current limitations, this block cannot be unset. To remove it once it has been set, the resource needs to be recreated.

  -> The `load_balancer_configuration` block contains:

    - `load_balancer_id` - (Required) The ID of the load balancer.

    - `backends` - (Required) The list of load balancer backend configurations.

      -> The `backends` block contains:

        - `backend_id` - (Required) The ID of the load balancer backend.
        - `address_family` - (Required) The IP address family (IPv4 or IPv6).
        - `private_network_id` - (Optional) The ID of the private network.

    - `auto_healing` - (Optional) The auto-healing configuration.

      -> The `auto_healing` block contains:

        - `enabled` - (Optional) Whether auto-healing is enabled.
        - `grace_period` - (Optional) The grace period for health checks.

- `zone` - (Defaults to provider `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the AutoScaling Group should be created.

- `project_id` - (Defaults to provider `project_id`) The ID of the project the AutoScaling Group is associated with.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the AutoScaling Group.

  ~> **Important:** AutoScaling Groups' IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

- `status` - The current status of the AutoScaling Group.

- `created_at` - The creation timestamp of the AutoScaling Group.

- `updated_at` - The last update timestamp of the AutoScaling Group.


## Import

AutoScaling Groups can be imported using their `id`:

```bash
terraform import scaleway_autoscaling_group.main <group_id>
```
