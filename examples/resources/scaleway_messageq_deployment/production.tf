resource "scaleway_messageq_deployment" "prod" {
  name       = "orders-messageq"
  version    = "4.0"
  node_count = 3
  node_type  = "MESSAGEQ-DEDICATED-4C-16G"
  user_name  = "admin"
  password   = var.admin_password

  volume {
    type       = "sbs_15k"
    size_in_gb = 20
  }

  tags = ["env=prod", "team=platform"]
}
