### Basic

# Find acls that share the same frontend ID
data "scaleway_lb_acls" "byFrontID" {
  frontend_id = scaleway_lb_frontend.frt01.id
}
# Find acls by frontend ID and name
data "scaleway_lb_acls" "byFrontID_and_name" {
  frontend_id = scaleway_lb_frontend.frt01.id
  name        = "tf-acls-datasource"
}
