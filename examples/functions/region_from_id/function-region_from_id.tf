# terraform block required for provider function to be found
terraform {
  required_providers {
    scaleway = {
      source = "scaleway/scaleway"
    }
  }
}

resource "scaleway_secret" "main" {
  name = "terraform_test_region_from_id"
}

output "region" {
  value = provider::scaleway::region_from_id(scaleway_secret.main.id)
}
