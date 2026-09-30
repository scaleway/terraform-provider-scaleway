### Basic

# Find frontends that share the same LB ID
data "scaleway_lb_frontends" "byLBID" {
  lb_id = scaleway_lb.lb01.id
}
# Find frontends by LB ID and name
data "scaleway_lb_frontends" "byLBID_and_name" {
  lb_id = scaleway_lb.lb01.id
  name  = "tf-frontend-datasource"
}
