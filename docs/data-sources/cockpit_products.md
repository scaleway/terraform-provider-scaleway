---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_products"
---

# Data Source: scaleway_cockpit_products

Gets the list of Cockpit products available for use in [`scaleway_cockpit_exporter`](../resources/cockpit_exporter.md) `exported_products`.

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) for more information.

## Example Usage

### List product names for exported_products

```terraform
data "scaleway_cockpit_products" "main" {}

resource "scaleway_cockpit_exporter" "main" {
  project_id        = data.scaleway_account_project.project.id
  datasource_id     = data.scaleway_cockpit_sources.scaleway_metrics.sources[0].id
  name              = "my-exporter"
  exported_products = data.scaleway_cockpit_products.main.names
}
```

### List products with details

```terraform
data "scaleway_cockpit_products" "main" {}

output "product_names" {
  value = [for p in data.scaleway_cockpit_products.main.products : p.name]
}

output "product_display_names" {
  value = [for p in data.scaleway_cockpit_products.main.products : p.display_name]
}
```

## Argument Reference

- `region` - (Optional) The [region](../guides/regions_and_zones.md#regions). Defaults to the provider region.

## Attributes Reference

- `products` - List of Cockpit products with details.
  - `name` - Product name to use in `exported_products` (e.g. `cockpit`, `lb`, `object-storage`).
  - `display_name` - Human-readable display name of the product.
  - `family_name` - Product family name.
- `names` - List of product names for use in `scaleway_cockpit_exporter.exported_products`.
