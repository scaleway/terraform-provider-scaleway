### Deploy your own model on your managed inference

resource "scaleway_inference_model" "my_model" {
  name   = "my-awesome-model"
  url    = "https://huggingface.co/agentica-org/DeepCoder-14B-Preview"
  secret = "my-secret-token"
}

resource "scaleway_inference_deployment" "my_deployment" {
  name      = "test-inference-deployment-basic"
  node_type = "H100" # replace with your node type
  model_id  = scaleway_inference_model.my_model.id

  public_endpoint {
    is_enabled = true
  }

  accept_eula = true
}
