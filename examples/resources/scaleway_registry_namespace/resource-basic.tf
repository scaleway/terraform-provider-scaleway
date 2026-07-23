### Basic

resource "scaleway_registry_namespace" "main" {
  name        = "main-cr"
  description = "Main container registry"
  is_public   = false
}
