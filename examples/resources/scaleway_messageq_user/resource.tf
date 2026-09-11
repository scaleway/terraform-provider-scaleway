resource "scaleway_messageq_deployment" "main" {
  name       = "my-messageq"
  version    = "4.0"
  node_count = 1
  node_type  = "MESSAGEQ-SHARED-2C-8G"
  user_name  = "admin"
  password   = "ThisIsASecurePassword123!"

  volume {
    type       = "sbs_5k"
    size_in_gb = 5
  }
}

resource "scaleway_messageq_user" "app" {
  deployment_id = scaleway_messageq_deployment.main.id
  name          = "orders-app"
  password      = var.app_password
}
