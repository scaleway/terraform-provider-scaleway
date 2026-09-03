### Create a secret

resource "scaleway_secret" "main" {
  name        = "foo"
  description = "barr"
  tags        = ["foo", "terraform"]
}
