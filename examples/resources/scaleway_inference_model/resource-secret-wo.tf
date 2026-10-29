### Create a model using your model's secret token without storing it in the state

resource "scaleway_inference_model" "my_model_wo" {
  name              = "my-awesome-model-wo"
  url               = "https://huggingface.co/agentica-org/DeepCoder-14B-Preview"
  secret_wo         = "my-secret-token"
  secret_wo_version = 1
}
