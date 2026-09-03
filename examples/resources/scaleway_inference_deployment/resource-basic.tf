### Basic

data "scaleway_inference_model" "my_model" {
  name = "meta/llama-3.1-8b-instruct:fp8"
}

resource "scaleway_inference_deployment" "deployment" {
  name       = "tf-inference-deployment"
  node_type  = "L4"
  model_name = data.scaleway_inference_model.my_model.id
  public_endpoint {
    is_enabled = true
  }
  accept_eula = true
}
