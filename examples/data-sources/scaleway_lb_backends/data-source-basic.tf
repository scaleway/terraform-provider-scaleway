### Basic

# Find backends that share the same LB ID
data "scaleway_lb_backends" "byLBID" {
  lb_id = scaleway_lb.lb01.id
}
# Find backends by LB ID and name
data "scaleway_lb_backends" "byLBID_and_name" {
  lb_id = scaleway_lb.lb01.id
  name  = "tf-backend-datasource"
}
