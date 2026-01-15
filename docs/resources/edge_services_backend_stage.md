---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_backend_stage"
---

# Resource: scaleway_edge_services_backend_stage

Creates and manages Scaleway Edge Services Backend Stages.

## Example Usage

### With object backend

```terraform
resource "scaleway_object_bucket" "main" {
  name = "my-bucket-name"
  tags = {
    foo = "bar"
  }
}

resource "scaleway_edge_services_pipeline" "main" {
  name = "my-pipeline"
}

resource "scaleway_edge_services_backend_stage" "main" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  s3_backend_config {
    bucket_name   = scaleway_object_bucket.main.name
    bucket_region = "fr-par"
  }
}
```

### With LB backend

```terraform
resource "scaleway_lb" "main" {
  ip_ids = [scaleway_lb_ip.main.id]
  zone   = "fr-par-1"
  type   = "LB-S"
}

resource "scaleway_lb_frontend" "main" {
  lb_id        = scaleway_lb.main.id
  backend_id   = scaleway_lb_backend.main.id
  name         = "frontend01"
  inbound_port = "443"
  certificate_ids = [
    scaleway_lb_certificate.cert01.id,
  ]
}

resource "scaleway_edge_services_pipeline" "main" {
  name = "my-pipeline"
}

resource "scaleway_edge_services_backend_stage" "main" {
  pipeline_id = scaleway_edge_services_pipeline.main.id
  lb_backend_config {
    lb_config {
      id          = scaleway_lb.main.id
      frontend_id = scaleway_lb_frontend.id
      is_ssl      = true
      zone        = "fr-par-1"
    }
  }
}
```

## Argument Reference

- `pipeline_id` - (Required) The ID of the pipeline.
- `s3_backend_config` - (Optional) The Scaleway Object Storage origin bucket (S3) linked to the backend stage.
    - `bucket_name` - The name of the Bucket.
    - `bucket_region` - The region of the Bucket.
    - `is_website` - Defines whether the bucket website feature is enabled.
- `lb_backend_config` - (Optional) The Scaleway Load Balancer linked to the backend stage.
    - `lb_config` - The Load Balancer config.
        - `id` - The ID of the Load Balancer.
        - `frontend_id` - The ID of the frontend.
        - `is_ssl` - Defines whether the Load Balancer's frontend handles SSL connections.
        - `domain_name` - The Fully Qualified Domain Name (in the format subdomain.example.com) to use in HTTP requests sent towards your Load Balancer.
        - `has_websocket` - Defines whether to forward websocket requests to the load balancer.
        - `zone` - (Defaults to [provider](../index.md#zone) `zone`) The [zone](../guides/regions_and_zones.md#zones) of the Load Balancer.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the backend stage is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the backend stage (UUID format).
- `created_at` - The date and time of the creation of the backend stage.
- `updated_at` - The date and time of the last update of the backend stage.

## Import

Backend stages can be imported using the `{id}`, e.g.

```bash
terraform import scaleway_edge_services_backend_stage.basic 11111111-1111-1111-1111-111111111111
```
