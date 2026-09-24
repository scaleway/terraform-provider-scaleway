### Filter products by family

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
