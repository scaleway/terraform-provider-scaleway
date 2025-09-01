---
subcategory: "Inference"
page_title: "Scaleway: scaleway_inference_model"
---

# scaleway_inference_model

The `scaleway_inference_model` data source allows you to retrieve information about an inference model available in the Scaleway Inference API, either by providing the model's `name` or its `model_id`.

## Example Usage

### Basic

```hcl
data "scaleway_inference_model" "my_model" {
  name = "meta/llama-3.1-8b-instruct:fp8"
}
```

## Argument Reference

You must provide either name or model_id, but not both.

- `name` (Optional, Conflicts with model_id) The fully qualified name of the model to look up (e.g., "meta/llama-3.1-8b-instruct:fp8"). The provider will search for a model with an exact name match in the selected region and project.
- `model_id` (Optional, Conflicts with name) The ID of the model to retrieve. Must be a valid UUID with locality (i.e., Scaleway's zoned UUID format).
- `project_id` (Optional) The project ID to use when listing models. If not provided, the provider default project is used.
- `region` (Optional) The region where the model is hosted. If not set, the provider default region is used.

## Attributes Reference

In addition to the input arguments above, the following attributes are exported:

- `id` - The unique identifier of the model.
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