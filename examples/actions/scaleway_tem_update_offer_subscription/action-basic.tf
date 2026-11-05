action "scaleway_tem_update_offer_subscription" "main" {
  config {
    project_id = "11111111-1111-1111-1111-111111111111"
    offer_name = "essential"
  }
}

resource "scaleway_tem_domain" "main" {
  name       = "example.com"
  accept_tos = true
  project_id = "11111111-1111-1111-1111-111111111111"

  lifecycle {
    action_trigger {
      events  = [before_create]
      actions = [action.scaleway_tem_update_offer_subscription.main]
    }
  }
}
