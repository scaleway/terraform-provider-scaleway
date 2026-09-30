### Use with cockpit_exporter

data "scaleway_cockpit_products" "main" {
  region = "fr-par"
}

resource "scaleway_cockpit_exporter" "main" {
  name              = "my-exporter"
  region            = "fr-par"
  exported_products = data.scaleway_cockpit_products.main.names
}
