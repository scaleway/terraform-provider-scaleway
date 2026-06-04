### Basic creation of an inference model

resource "scaleway_inference_model" "test" {
  name   = "my-awesome-model"
  url    = "https://huggingface.co/agentica-org/DeepCoder-14B-Preview"
  secret = "my-secret-token"
}
