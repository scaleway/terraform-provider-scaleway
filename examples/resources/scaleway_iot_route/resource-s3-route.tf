### S3 Route

resource "scaleway_iot_route" "main" {
  name   = "main"
  hub_id = scaleway_iot_hub.main.id
  topic  = "#"
  s3 {
    bucket_region = scaleway_object_bucket.main.region
    bucket_name   = scaleway_object_bucket.main.name
    object_prefix = "foo"
    strategy      = "per_topic"
  }
}

resource "scaleway_iot_hub" "main" {
  name         = "main"
  product_plan = "plan_shared"
}

resource "scaleway_object_bucket" "main" {
  region = "fr-par"
  name   = "my_awesome-bucket"
}
