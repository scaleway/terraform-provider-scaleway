# List all PoPs in a region
data "scaleway_interlink_pops" "all" {
  region = "fr-par"
}

# List PoPs with a specific hosting provider name
data "scaleway_interlink_pops" "by_hosting_provider_name" {
  hosting_provider_name = "OpCore"
}

# List PoPs with dedicated connections available
data "scaleway_interlink_pops" "dedicated" {
  dedicated_available = true
}

