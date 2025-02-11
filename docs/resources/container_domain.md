---
subcategory: "Containers"
page_title: "Scaleway: scaleway_container_domain"
---

# Resource: scaleway_container_domain

The `scaleway_container_domain` resource allows you to create and manage domain name bindings for Scaleway [Serverless Containers](https://www.scaleway.com/en/docs/serverless/containers/).

Refer to the Containers domain [documentation](https://www.scaleway.com/en/docs/serverless-containers/how-to/add-a-custom-domain-to-a-container/) and the [API documentation](https://www.scaleway.com/en/developers/api/serverless-containers/#path-domains-list-all-domain-name-bindings) for more information.

## Example Usage

The commands below shows how to bind a custom domain name to a container.

### Simple

```terraform
resource scaleway_container app {}

resource scaleway_container_domain "app" {
  container_id = scaleway_container.app.id
  hostname = "container.domain.tld"
}
```

### Complete example with domain

```terraform
resource scaleway_container_namespace main {
    name = "my-ns-test"
    description = "test container"
}

resource scaleway_container app {
    name = "app"
    namespace_id = scaleway_container_namespace.main.id
    registry_image = "${scaleway_container_namespace.main.registry_endpoint}/nginx:alpine"
    port = 80
    cpu_limit = 140
    memory_limit = 256
    min_scale = 1
    max_scale = 1
    timeout = 600
    max_concurrency = 80
    privacy = "public"
    protocol = "http1"
    deploy = true
}

resource scaleway_domain_record "app" {
  dns_zone = "domain.tld"
  name     = "subdomain"
  type     = "CNAME"
  data     = "${scaleway_container.app.domain_name}." // Trailing dot is important in CNAME
  ttl      = 3600
}

resource scaleway_container_domain "app" {
  container_id = scaleway_container.app.id
  hostname = "${scaleway_domain_record.app.name}.${scaleway_domain_record.app.dns_zone}"
}
```

## Argument Reference

The following arguments are required:

- `hostname` - (Required) The hostname with a CNAME record.

- `container_id` - (Required) The unique identifier of the container.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the container exists.

## Attributes Reference

The `scaleway_container_domain` resource exports certain attributes once the container domain name has been retrieved. These attributes can be referenced in other parts of your Terraform configuration.

- `id` - The unique identifier of the container domain.

~> **Important:** Container domain IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`

- `url` - The URL used to query the container.


## Import

Container domain binding can be imported using `{region}/{id}`, as shown below:

```bash
terraform import scaleway_container_domain.main fr-par/11111111-1111-1111-1111-111111111111
```
