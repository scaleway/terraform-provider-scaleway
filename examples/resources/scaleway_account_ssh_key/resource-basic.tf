### Basic

resource "scaleway_account_ssh_key" "main" {
  name       = "main"
  public_key = "<YOUR-PUBLIC-SSH-KEY>"
}
