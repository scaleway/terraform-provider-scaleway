### Basic

resource "scaleway_apple_silicon_runner" "main" {
  name        = "my-github-runner"
  ci_provider = "github"
  url         = "https://github.com/my-org/my-repo"
  token       = "my-token"
}
