### Basic

# Get api key infos by id (access_key)
data "scaleway_iam_api_key" "main" {
  access_key = "SCWABCDEFGHIJKLMNOPQ"
}
