### Create a policy with a particular condition

resource "scaleway_iam_policy" "main" {
  name         = "tf_tests_policy_condition"
  no_principal = true
  rule {
    organization_id      = "%s"
    permission_set_names = ["AllProductsFullAccess"]
    condition            = "request.user_agent == 'My User Agent'"
  }
}
