### Basic

# Find routes that share the same frontend ID
data "scaleway_lb_routes" "by_frontendID" {
  frontend_id = scaleway_lb_frontend.frt01.id
}
# Find routes by frontend ID and zone
data "scaleway_lb_routes" "my_key" {
  frontend_id = "11111111-1111-1111-1111-111111111111"
  zone        = "fr-par-2"
}
