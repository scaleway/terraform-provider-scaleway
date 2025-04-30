---
subcategory: "Inference"
page_title: "Scaleway: scaleway_inference_model"
---

# Resource: scaleway_inference_model

The scaleway_inference_model resource allows you to upload and manage custom inference models in the Scaleway Inference ecosystem. Once registered, a custom model can be used in any scaleway_inference_deployment resource.

## Example Usage

### Basic

```terraform
resource "scaleway_inference_model" "test" {
  name = "my-awesome-model"
  url = "https://huggingface.co/agentica-org/DeepCoder-14B-Preview"
  secret = "my-secret-token"
}
```

### Deploy your own model on your managed inference

```terraform
resource "scaleway_inference_model" "my_model" {
  name = "my-awesome-model"
  url = "https://huggingface.co/agentica-org/DeepCoder-14B-Preview"
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
```

## Argument Reference

- `name` - (Required) The name of the custom model. This must be unique within the project.
- `url` - (Required) The HTTPS source URL from which the model will be downloaded. This is typically a Hugging Face repository URL (e.g., https://huggingface.co/agentica-org/DeepCoder-14B-Preview). The URL must be publicly accessible or require valid credentials via `secret`
- `secret` - (Optional, Sensitive) Authentication token used to pull the model from a private or gated URL (e.g., a Hugging Face access token with read permission).
- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the deployment is created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the deployment is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier of the custom model.
- `tags` - Tags associated with the model.
- `status` - The current status of the model (e.g., ready, error, etc.).
- `description` - A textual description of the model (if available).
- `has_eula` - Whether the model requires end-user license agreement acceptance before use.
- `parameter_size_bits` - Size, in bits, of the model parameters.
- `size_bytes` - Total size, in bytes, of the model archive.
- `nodes_support` - List of supported node types and their quantization options. Each entry contains:
        - `node_type_name` - The type of node supported.
        - `quantization` - A list of supported quantization options, including:
            - `quantization_bits` -  Number of bits used for quantization (e.g., 8, 16).
            - `allowed` - Whether this quantization is allowed.
            - `max_context_size` - Maximum context length supported by this quantization.

## Import

Custom models can be imported using, `{region}/{id}`, as shown below:

```bash
terraform import scaleway_inference_model.my_model fr-par/11111111-1111-1111-1111-111111111111
```