### Programmatic access with the Grafana provider

variable "scaleway_secret_key" {
  description = "Scaleway IAM secret key used for both the Scaleway and Grafana providers"
  type        = string
  sensitive   = true
}

data "scaleway_cockpit_grafana" "main" {
  project_id = scaleway_account_project.project.id
}

provider "grafana" {
  url  = data.scaleway_cockpit_grafana.main.grafana_url
  auth = "anonymous"

  http_headers = {
    "X-Auth-Token" = var.scaleway_secret_key
  }
}
