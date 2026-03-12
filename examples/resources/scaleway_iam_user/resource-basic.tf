### Basic IAM user creation

resource "scaleway_iam_user" "user" {
  email      = "foo@test.com"
  tags       = ["test-tag"]
  username   = "foo"
  first_name = "Foo"
  last_name  = "Bar"
}
