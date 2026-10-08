data "scaleway_cockpit_rules_count" "main" {
  region     = "fr-par"
  project_id = "11111111-1111-1111-1111-111111111111"
}

output "custom_rules" {
  value = data.scaleway_cockpit_rules_count.main.custom_rules_count
}
