### Basic

resource "scaleway_iam_group" "basic" {
  name            = "iam_group_basic"
  description     = "basic description"
  application_ids = []
  user_ids        = []
}
