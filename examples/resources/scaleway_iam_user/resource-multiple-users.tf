### Multiple IAM user creation

locals {
  users = [
    {
      email    = "test@test.com"
      username = "test"
    },
    {
      email    = "test2@test.com"
      username = "test2"
    }
  ]
}

resource "scaleway_iam_user" "users" {
  count    = length(local.users)
  email    = local.users[count.index].email
  username = local.users[count.index].username
}
