# List servers filtered by name and tag on a specific zone across all projects
list "scaleway_instance_server" "by_name_and_tag" {
  provider = scaleway

  config {
    zones       = ["fr-par-1"]
    project_ids = ["*"]
    name        = "the-exact-name"
    tags        = ["a-specific-tag"]
  }
}