terraform {
  required_providers {
    scaleway = {
      source = "scaleway/scaleway"
    }
  }
  required_version = ">= 0.14"
}

resource "scaleway_vpc" "main" {
  name = "main"
}

resource "scaleway_vpc_private_network" "main" {
  name = "main"
  vpc_id = scaleway_vpc.main.id
}

resource "scaleway_container_namespace" "main" {
  name = "main"
}

resource "scaleway_container" "public" {
  name = "public"
  namespace_id = scaleway_container_namespace.main.id
  image = "rg.fr-par.scw.cloud/serverless-catalog/sample-website:latest"
  enable_default_public_endpoint = true
}

output "public-endpoint" {
  value = scaleway_container.public.public_endpoint
}

resource "scaleway_container" "private" {
  name = "private"
  namespace_id = scaleway_container_namespace.main.id
  image = "rg.fr-par.scw.cloud/serverless-catalog/sample-website:latest"
  enable_default_public_endpoint = false
  enable_private_endpoint = true
  private_network_id = scaleway_vpc_private_network.main.id
}

output "private-endpoint" {
  value = scaleway_container.private.private_endpoint
}
