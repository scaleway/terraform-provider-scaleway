---
subcategory: "Instances"
page_title: "Scaleway: scaleway_instance_template"
---

# Resource: scaleway_instance_template

Creates and manages Scaleway Instance Templates.
For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/instance/templates).

## Example Usage

### Basic

```terraform
resource "scaleway_instance_template" "main" {
  name = "instance-template-basic"
  tags = ["terraform-examples", "scaleway_instance_template", "basic"]

  server_type       = "PRO2-M"
  server_tags       = ["created-from-template"]
  public_ipv4_count = 1
  public_ipv6_count = 3
}
```

### With volumes

```terraform
resource "scaleway_block_volume" "sbs" {
  size_in_gb = 15
  iops       = 5000
}
resource "scaleway_block_snapshot" "snap" {
  volume_id = scaleway_block_volume.sbs.id
}

resource "scaleway_instance_template" "main" {
  name        = "instance-template-with-volumes"
  server_type = "GP1-L"

  volumes = [
    {
      volume_type = "l_ssd"
      image_label = "debian_trixie"
      size_in_gb  = 25
      tags        = ["local"]
      name        = "root-volume"
    },
    {
      volume_type      = "sbs"
      base_snapshot_id = scaleway_block_snapshot.snap.id
      size_in_gb       = 30
      perf_iops        = 15000
    }
  ]
}
```

### With scratch storage

```terraform
resource "scaleway_instance_template" "main" {
  name        = "instance-template-with-scratch-storage"
  server_type = "L40S-1-48G"
  zone        = "fr-par-2"

  volumes = [
    {
      volume_type = "sbs"
      image_label = "ubuntu_noble_gpu_os_13_nvidia"
      size_in_gb  = 30
      perf_iops   = 15000
    },
    {
      volume_type = "scratch"
      size_in_gb  = 300
    }
  ]
}
```

### Windows

```terraform
resource "scaleway_iam_ssh_key" "key" {
  name       = "instance-tmpl-admin-ssh-key"
  public_key = file("~/.ssh/id_rsa.pub")
}

resource "scaleway_instance_template" "main" {
  name        = "instance-template-windows"
  server_type = "POP2-4C-16G-WIN"

  windows_rdp_ssh_key_id = scaleway_iam_ssh_key.key.id
}
```

### With additional resources

```terraform
# Security Group
resource "scaleway_instance_security_group" "sg" {}
# Placement Group
resource "scaleway_instance_placement_group" "pg" {}
# VPC + Private Network
resource "scaleway_vpc" "vpc" {}
resource "scaleway_vpc_private_network" "pn" {
  vpc_id = scaleway_vpc.vpc.id
}
# Filesystem
resource "scaleway_file_filesystem" "fs" {
  size_in_gb = 25
}

resource "scaleway_instance_template" "main" {
  name        = "instance-template-additional-resources"
  server_type = "PRO2-M"

  security_group_id  = scaleway_instance_security_group.sg.id
  placement_group_id = scaleway_instance_placement_group.pg.id
  private_networks   = [scaleway_vpc_private_network.pn.id]
  filesystem_ids     = [scaleway_file_filesystem.fs.id]
}
```


## Argument Reference

The following arguments are supported:

- `server_type` - (Required) The commercial type of the server defined in the template.

- `name` - (Optional) The name of the template. If not provided it will be randomly generated.

- `tags` - (Optional) A list of tags to apply to the template.

- `server_tags` - (Optional) A list of tags to apply to the servers created from the template.

- `public_ipv4_count` - (Defaults to `0`) The number of public IPv4 to attach to the servers created using the template.

- `public_ipv6_count` - (Defaults to `0`) The number of public IPv6 to attach to the servers created using the template.

- `volumes` - (Optional) The list of specs describing the volumes to attach to the servers created using the template.

  -> The `volumes` block contains :

    - `volume_type` - (Required) The type of the volume.

    - `size_in_gb` - (Required)

    - `name` - (Optional) The name of volume.

    - `tags` - (Optional) The tags associated with the volume.

    - `base_snapshot_id` - (Optional) The ID of the base snapshot for the volume.

      ~> **Important:** Only one of `base_snapshot_id` and `image_label` can be set.

    - `image_label` - (Optional) The label of the image used as base for the volume.

    - `perf_iops` - (Optional) The performance IOPS of the volume, required for `sbs` type volumes.

- `security_group_id` - (Optional) The ID of the security group to attach to the servers created using the template.

- `placement_group_id` - (Optional) The ID of the placement group to attach to the servers created using the template.

- `private_networks` - (Optional) - The IDs of the private networks to attach to the servers created using the template.

- `filesystem_ids` - (Optional) - The IDs of the filesystems to attach to the servers created using the template.

- `windows_rdp_ssh_key_id` - (Optional) The ID of the IAM SSH key used to encrypt the initial admin password on a Windows server. This will be repeated on all servers created using the template.

- `zone` - (Defaults to provider `zone`) The [zone](../guides/regions_and_zones.md#zones) in which the template should be created.

- `project_id` - (Defaults to provider `project_id`) The ID of the project the template is associated with.


## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the template.

  ~> **Important:** Instance templates' IDs are [zoned](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{zone}/{id}`, e.g. `fr-par-1/11111111-1111-1111-1111-111111111111`

- `created_at` - The creation timestamp of the Instance template.

- `updated_at` - The last update timestamp of the Instance template.


## Import

Templates can be imported using the `{zone}/{id}`, e.g.

```bash
terraform import scaleway_instance_template.main fr-par-1/11111111-1111-1111-1111-111111111111
```
