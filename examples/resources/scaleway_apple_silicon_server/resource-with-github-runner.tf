### With `github` runner

data "scaleway_apple_silicon_os" "by_name" {
  name = "devos-sequoia-15.6"
}

resource "scaleway_apple_silicon_runner" "main" {
  name        = "TestAccRunnerGithub"
  ci_provider = "github"
  url         = "https://github.com/my-repo-url"
  token       = "MY_GITHUB_RUNNER_TOKEN"
}

resource "scaleway_apple_silicon_server" "main" {
  name             = "TestAccServerRunner"
  type             = "M2-L"
  public_bandwidth = 1000000000
  os_id            = data.scaleway_apple_silicon_os.by_name.id
  runner_ids       = [scaleway_apple_silicon_runner.main.id]
}
