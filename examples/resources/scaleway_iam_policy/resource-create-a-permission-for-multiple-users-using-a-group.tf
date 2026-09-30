### Create a permission for multiple users using a group

locals {
  users = [
    "user1@mail.com",
    "user2@mail.com",
  ]
  project_name = "default"
}

data "scaleway_account_project" "project" {
  name = local.project_name
}

data "scaleway_iam_user" "users" {
  for_each = toset(local.users)
  email    = each.value
}

resource "scaleway_iam_group" "with_users" {
  name     = "developers"
  user_ids = [for user in data.scaleway_iam_user.users : user.id]
}

resource "scaleway_iam_policy" "iam_tf_storage_policy" {
  name     = "developers permissions"
  group_id = scaleway_iam_group.with_users.id
  rule {
    project_ids          = [data.scaleway_account_project.project.id]
    permission_set_names = ["InstancesReadOnly"]
  }
}
