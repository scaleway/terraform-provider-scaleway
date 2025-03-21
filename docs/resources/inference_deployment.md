---
subcategory: "Inference"
page_title: "Scaleway: scaleway_inference_deployment"
---

# Resource: scaleway_inference_deployment

Creates and manages Scaleway Managed Inference deployments.
For more information, see the [API documentation](https://www.scaleway.com/en/developers/api/inference/).

## Example Usage

### Basic

```terraform
resource "scaleway_inference_deployment" "deployment" {
  name = "tf-inference-deployment"
  node_type = "L4"
  model_name = "meta/llama-3.1-8b-instruct:fp8"
  public_endpoint {
    is_enabled = true
  }
  accept_eula = true
}
```

## Argument Reference

- `model_name` - (Required) The model name to use for the deployment. Model names can be found in Console or using Scaleway's CLI (`scw inference model list`)
- `node_type` - (Required) The node type to use for the deployment. Node types can be found using Scaleway's CLI (`scw inference node-type list`)
- `name` - (Optional) The deployment name.
- `accept_eula` - (Optional) Some models (e.g Meta Llama) require end-user license agreements. Set `true` to accept.
- `tags` - (Optional) The tags associated with the deployment.
- `min_size` - (Optional) The minimum size of the pool.
- `max_size` - (Optional) The maximum size of the pool.
- `private_endpoint` - (Optional) Configuration of the deployment's private endpoint.
    - `private_network_id` - (Optional) The ID of the private network to use.
    - `disable_auth` - (Optional) Disable the authentication on the endpoint.
- `public_endpoint` - (Optional) Configuration of the deployment's public endpoint.
    - `is_enabled` - (Optional) Enable or disable public endpoint.
    - `disable_auth` - (Optional) Disable the authentication on the endpoint.

- `region` - (Defaults to [provider](../index.md#region) `region`) The [region](../guides/regions_and_zones.md#regions) in which the deployment is created.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the deployment is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the deployment.
- `model_id` - The model id used for the deployment.
- `size` - The size of the pool.
- `status` - The status of the deployment.
- `created_at` - The date and time of the creation of the deployment.
- `updated_at` - The date and time of the last update of the deployment.
- `private_endpoint` - Private endpoint's attributes.
    - `id` - (Optional) The id of the private endpoint.
    - `url` - (Optional) The URL of the endpoint.
- `public_endpoint` - (Optional) Public endpoint's attributes.
    - `id` - (Optional) The id of the public endpoint.
    - `url` - (Optional) The URL of the endpoint.

~> **Important:** Deployments' IDs are [regional](../guides/regions_and_zones.md#resource-ids), which means they are of the form `{region}/{id}`, e.g. `fr-par/11111111-1111-1111-1111-111111111111`.


## Import

Functions can be imported using, `{region}/{id}`, as shown below:

```bash
terraform import scaleway_inference_deployment.deployment fr-par/11111111-1111-1111-1111-111111111111
```
