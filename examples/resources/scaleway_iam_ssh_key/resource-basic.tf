### Basic

resource "scaleway_iam_ssh_key" "main" {
  name       = "main"
  public_key = "<YOUR-PUBLIC-SSH-KEY>"
}
