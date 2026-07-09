### Get all available products

data "scaleway_cockpit_products" "main" {
  region = "fr-par"
}

output "available_products" {
  value = data.scaleway_cockpit_products.main.products
}
