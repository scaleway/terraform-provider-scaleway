---
page_title: "Migrating instance volumes to SBS"
---

# Migration

This page describe how migrate your instance's volumes to SBS (Scaleway Block Storage).
You can migrate your instance's volumes if they are block volumes (`b_ssd`).

Migration of local volumes is not supported, you'll have to migrate your data manually.

## Migrate your implicit root_volume

Given your infrastructure with a server that has a root_volume that must be migrated:

```terraform
resource scaleway_instance_server "server" {
  type  = "PLAY2-PICO"
  image = "ubuntu_jammy"
  root_volume {
    volume_type = "b_ssd"
  }
}
```

In the previous snippet, the root_volume type is explicitly configured as a b_ssd, this configuration must be removed to prepare for migration.

```terraform
resource scaleway_instance_server "server" {
  type  = "PLAY2-PICO"
  image = "ubuntu_jammy"
}
```

You can now migrate your root_volume using Scaleway's CLI ([documentation](https://www.scaleway.com/en/docs/instances/how-to/migrate-volumes-snapshots-to-sbs/#migrating-an-existing-block-storage-volume-to-scaleway-block-storage-management)).
After migration, your terraform plan should be empty as the provider should have picked up that the volume is now managed by SBS API.

## Migrate your explicit volumes

Given your infrastructure with a server and explicit volumes.

```terraform
resource scaleway_instance_volume "root_volume" {
  size_in_gb = 20
  type = "b_ssd"
}

resource scaleway_instance_volume "volume" {
  size_in_gb = 20
  type = "b_ssd"
}

resource scaleway_instance_server "server" {
  type  = "PLAY2-PICO"
  root_volume {
    volume_id = scaleway_instance_volume.root_volume.id
  }

  additional_volume_ids = [scaleway_instance_volume.volume.id]
}
```

You can rely on terraform to realize migration, it will be done in 2 steps.
The first one will be a transitional state to migrate.
The `migrate_to_sbs` field in your instance_volume will prevent the volume state from being updated during our migration.


```terraform
resource scaleway_instance_volume "root_volume" {
  size_in_gb = 20
  type = "b_ssd"
  migrate_to_sbs = true # Mark migration to avoid failure
}

resource scaleway_block_volume "root_volume" {
  size_in_gb = 20
  iops = 5000 # b_ssd is a 5000 iops volume
  instance_volume_id = scaleway_instance_volume.root_volume.id # block resource will handle migration
}

resource scaleway_instance_volume "volume" {
  size_in_gb = 20
  type = "b_ssd"
  migrate_to_sbs = true # Mark migration to avoid failure
}

resource scaleway_block_volume "volume" {
  size_in_gb = 20
  iops = 5000 # b_ssd is a 5000 iops volume
  instance_volume_id = scaleway_instance_volume.volume.id # block resource will handle migration
}


resource scaleway_instance_server "server" {
  type  = "PLAY2-PICO"
  root_volume {
    volume_id = scaleway_block_volume.root_volume.id # Start using your new resource
  }

  additional_volume_ids = [scaleway_block_volume.volume.id] # Start using your new resource
}
```

The first migration step should have created your new block volumes and updated the old instance_volume resources.
You can now remove the old instance resources for the final step.
To be sure, the resource `scaleway_instance_volume` is not capable of deleting a volume that is on SBS API, you can check using Console or CLI that your volume successfully migrated before applying the final step.


```terraform
resource scaleway_block_volume "root_volume" {
  size_in_gb = 20
  iops = 5000
}

resource scaleway_block_volume "volume" {
  size_in_gb = 20
  iops = 5000
}

resource scaleway_instance_server "server" {
  type  = "PLAY2-PICO"
  root_volume {
    volume_id = scaleway_block_volume.root_volume.id
  }

  additional_volume_ids = [scaleway_block_volume.volume.id]
}
```