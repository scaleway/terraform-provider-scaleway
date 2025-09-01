---
page_title: "Migrating the management of Instance Block Storage volumes to SBS"
---

# Migration

This page describes how to migrate the management of your Instance's volumes from Block Storage legacy to SBS (Scaleway Block Storage).
This documentation **only applies if you have created Block Storage legacy volumes** (`b_ssd`).

Migration of local volumes is not supported, you will have to migrate your data manually.

Find out about the advantages of migrating from the Instance API to the Block Storage API for managing block volumes in the [dedicated documentation](https://www.scaleway.com/en/docs/block-storage/reference-content/advantages-migrating-to-sbs/).

## Migrate your implicit root_volume

If your infrastructure includes a server with a root volume that must be migrated:

```terraform
resource scaleway_instance_server "server" {
  type  = "PLAY2-PICO"
  image = "ubuntu_jammy"
  root_volume {
    volume_type = "b_ssd"
  }
}
```

In the snippet above, the `root_volume` type is explicitly configured as a `b_ssd`. This configuration must be removed to prepare for migration.

```terraform
resource scaleway_instance_server "server" {
  type  = "PLAY2-PICO"
  image = "ubuntu_jammy"
}
```

You can now migrate your root_volume using the [Scaleway CLI documentation](https://www.scaleway.com/en/docs/instances/how-to/migrate-volumes-snapshots-to-sbs/#migrating-an-existing-block-storage-volume-to-scaleway-block-storage-management).
After migration, the output when running `terraform plan` should be empty as the provider should have picked up that the volume is now managed by the Scaleway Block Storage API.

## Migrate your explicit volumes

If your infrastructure includes servers and explicit volumes.

```terraform
resource scaleway_instance_volume "root_volume" {
  size_in_gb = 20
  type       = "b_ssd"
}

resource scaleway_instance_volume "volume" {
  size_in_gb = 20
  type       = "b_ssd"
}

resource scaleway_instance_server "server" {
  type = "PLAY2-PICO"
  root_volume {
    volume_id = scaleway_instance_volume.root_volume.id
  }

  additional_volume_ids = [scaleway_instance_volume.volume.id]
}
```

You can rely on Terraform to perform the migration which will be done in 2 steps.
The first one will be a transitional state to migrate and is described in the next snippet.
In this snippet, the `migrate_to_sbs` field will prevent the old volume state from being updated during the migration.


```terraform
resource scaleway_instance_volume "root_volume" {
  size_in_gb     = 20
  type           = "b_ssd"
  migrate_to_sbs = true # Mark migration to avoid failure
}

resource scaleway_block_volume "root_volume" {
  size_in_gb         = 20
  iops               = 5000                                    # b_ssd is a 5000 iops volume
  instance_volume_id = scaleway_instance_volume.root_volume.id # block resource will handle migration
}

resource scaleway_instance_volume "volume" {
  size_in_gb     = 20
  type           = "b_ssd"
  migrate_to_sbs = true # Mark migration to avoid failure
}

resource scaleway_block_volume "volume" {
  size_in_gb         = 20
  iops               = 5000                               # b_ssd is a 5000 iops volume
  instance_volume_id = scaleway_instance_volume.volume.id # block resource will handle migration
}


resource scaleway_instance_server "server" {
  type = "PLAY2-PICO"
  root_volume {
    volume_id = scaleway_block_volume.root_volume.id # Start using your new resource
  }

  additional_volume_ids = [scaleway_block_volume.volume.id] # Start using your new resource
}
```

The first migration step should have created your new Block Storage volumes and updated the existing instance_volume resources.
After confirming the migration is successful, you must remove the old Instance's resources manually.
Terraform's scaleway_instance_volume resource cannot delete a volume that has been migrated to the Scaleway Block Storage API. Before applying the final step, check in the Scaleway [console](https://console.scaleway.com) or using the [CLI](https://cli.scaleway.com/block/#list-volumes) to confirm that your volume was successfully migrated.


```terraform
resource scaleway_block_volume "root_volume" {
  size_in_gb = 20
  iops       = 5000
}

resource scaleway_block_volume "volume" {
  size_in_gb = 20
  iops       = 5000
}

resource scaleway_instance_server "server" {
  type = "PLAY2-PICO"
  root_volume {
    volume_id = scaleway_block_volume.root_volume.id
  }

  additional_volume_ids = [scaleway_block_volume.volume.id]
}
```
