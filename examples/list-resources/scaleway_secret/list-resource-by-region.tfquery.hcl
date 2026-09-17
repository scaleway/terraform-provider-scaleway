// List secrets in specific regions
list "scaleway_secret" "by_region" {
  provider = scaleway

  config {
    regions = ["fr-par", "nl-ams"]
  }
}
