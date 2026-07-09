### Users membership

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

resource "scaleway_iam_group" "group" {
  name                = "my_group"
  external_membership = true
}

resource "scaleway_iam_group_membership" "members" {
  for_each = data.scaleway_iam_user.users
  group_id = scaleway_iam_group.group.id
  user_id  = each.value.id
}
