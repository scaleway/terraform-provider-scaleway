---
subcategory: "Edge Services"
page_title: "Scaleway: scaleway_edge_services_route_stage"
---

# Resource: scaleway_edge_services_route_stage

Creates and manages Scaleway Edge Services Route Stages.

## Example Usage

### Basic

```terraform
resource "scaleway_edge_services_route_stage" "main" {
  pipeline_id   = scaleway_edge_services_pipeline.main.id
  waf_stage_id  = scaleway_edge_services_waf_stage.waf.id
  after_position = 0

  rules {
    backend_stage_id = scaleway_edge_services_backend_stage.backend.id

    rule_http_match {
      method_filters = ["get", "post"]

      path_filter {
        path_filter_type = "regex"
        value           = ".*"
      }
    }
  }
}
```

## Argument Reference

- `pipeline_id` - (Required) The ID of the pipeline.
- `after_position` - (Optional) Add rules after the given position. Only one of `after_position` and `before_position` should be specified.
- `before_position` - (Optional) Add rules before the given position. Only one of `after_position` and `before_position` should be specified.
- `waf_stage_id` - (Optional) The ID of the WAF stage HTTP requests should be forwarded to when no rules are matched.
- `rules` - (Optional) The list of rules to be checked against every HTTP request. The first matching rule will forward the request to its specified backend stage. If no rules are matched, the request is forwarded to the WAF stage defined by `waf_stage_id`.
    - `backend_stage_id` (Required) The ID of the backend stage that requests matching the rule should be forwarded to.
    - `rule_http_match` (Optional) The rule condition to be matched. Requests matching the condition defined here will be directly forwarded to the backend specified by the `backend_stage_id` field. Requests that do not match will be checked by the next rule's condition.
        - `method_filters` (Optional) HTTP methods to filter for. A request using any of these methods will be considered to match the rule. Possible values are `get`, `post`, `put`, `patch`, `delete`, `head`, `options`. All methods will match if none is provided.
        - `path_filter` (Optional) HTTP URL path to filter for. A request whose path matches the given filter will be considered to match the rule. All paths will match if none is provided.
            - `path_filter_type` (Required) The type of filter to match for the HTTP URL path. For now, all path filters must be written in regex and use the `regex` type.
            - `value` (Required) The value to be matched for the HTTP URL path.
- `project_id` - (Defaults to [provider](../index.md#project_id) `project_id`) The ID of the project the route stage is associated with.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The ID of the route stage (UUID format).
- `created_at` - The date and time of the creation of the route stage.
- `updated_at` - The date and time of the last update of the route stage.

## Import

Route stages can be imported using the `{id}`, e.g.

```bash
$ terraform import scaleway_edge_services_route_stage.basic 11111111-1111-1111-1111-111111111111
```
