---
page_title: "scaleway_instance_image Data Source - terraform-provider-scaleway"
subcategory: ""
description: |-
  
---

# Data Source `scaleway_instance_image`





## Schema

### Optional

- **architecture** (String) Architecture of the desired image
- **id** (String) The ID of this resource.
- **image_id** (String) ID of the desired image
- **latest** (Boolean) Select most recent image if multiple match
- **name** (String) Exact name of the desired image
- **project_id** (String) The project_id you want to attach the resource to
- **zone** (String) The zone you want to attach the resource to

### Read-only

- **additional_volume_ids** (List of String) The additional volume IDs attached to the image
- **creation_date** (String) Date when the image was created
- **default_bootscript_id** (String) ID of the bootscript associated with this image
- **from_server_id** (String) ID of the server the image is originated from
- **modification_date** (String) Date when the image was updated
- **organization_id** (String) The organization_id you want to attach the resource to
- **public** (Boolean) Indication if the image is public
- **root_volume_id** (String) ID of the root volume associated with this image
- **state** (String) State of the image


