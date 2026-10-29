---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_products"
---

# Data Source: scaleway_cockpit_products

Gets the list of Cockpit products available for a specific region.

Use this data source to retrieve the list of products that can be exported via [`scaleway_cockpit_exporter`](../resources/cockpit_exporter.md). The products are returned with their machine-readable name, display name, and family name.

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Example Usage

### Get all available products

```terraform
data "scaleway_cockpit_products" "main" {
  region = "fr-par"
}

output "available_products" {
  value = data.scaleway_cockpit_products.main.products
}
```

### Use with cockpit_exporter

```terraform
data "scaleway_cockpit_products" "main" {
  region = "fr-par"
}

resource "scaleway_cockpit_exporter" "main" {
  name              = "my-exporter"
  region            = "fr-par"
  exported_products = data.scaleway_cockpit_products.main.names
}
```

### Filter products by family

```terraform
data "scaleway_cockpit_products" "main" {
  region = "fr-par"
}

locals {
  # Filter to only compute products
  compute_products = [
    for p in data.scaleway_cockpit_products.main.products
    : p.name
    if p.family_name == "Compute"
  ]
}

output "compute_product_names" {
  value = local.compute_products
}
```

## Argument Reference

- `region` - (Optional) The [region](../guides/regions_and_zones.md#regions) to query. Defaults to the region configured in the provider.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The region ID (same as `region`).
- `products` - List of available Cockpit products. (see [below](#nestedatt--products))
- `names` - List of product names that can be directly used in `scaleway_cockpit_exporter.exported_products`.

<a id="nestedatt--products"></a>
### Nested Schema for `products`

- `name` - The machine-readable product name to use in `exported_products` (e.g. `cockpit`, `lb`, `object-storage`).
- `display_name` - Human-readable display name of the product.
- `family_name` - Product family category (e.g. `Compute`, `Network`, `Storage`).
