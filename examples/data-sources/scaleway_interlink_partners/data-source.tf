# List all partners in a region
data "scaleway_interlink_partners" "all" {
  region = "fr-par"
}

# List partners available at specific PoPs
data "scaleway_interlink_partners" "at_pops" {
  pop_ids = [
    data.scaleway_interlink_pop.main.id,
  ]
}
