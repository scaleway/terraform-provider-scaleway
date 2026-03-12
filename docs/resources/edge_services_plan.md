---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_plan"
---

# Resource: scaleway_edge_services_plan

Creates and manages Scaleway Edge Services plans.

## Example Usage

### Basic

```terraform
resource "scaleway_edge_services_plan" "main" {
  name = "starter"
}
```

## Argument Reference

- `name` - (Optional) The name of the plan.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the plan is associated with.

## Attributes Reference

No additional attributes are exported.

## Import

Plans can be imported using `{project_id}/{plan_name}`, e.g.

```bash
terraform import scaleway_edge_services_plan.main 11111111-1111-1111-1111-111111111111/starter
```
