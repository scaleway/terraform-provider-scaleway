### With users

locals {
  users = toset([
    "user1@mail.com",
    "user2@mail.com"
  ])
}

data "scaleway_iam_user" "users" {
  for_each = local.users
  email    = each.value
}

resource "scaleway_iam_group" "with_users" {
  name            = "iam_group_with_app"
  application_ids = []
  user_ids        = [for user in data.scaleway_iam_user.users : user.id]
}
