---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_plan"
---
# scaleway_cockpit_plan

The `scaleway_cockpit_plan` data source is used to fetch details about a specific Scaleway Cockpit pricing plan. This information can then be used to configure resources like `scaleway_cockpit`.

Find out more about [pricing plans](https://console.scaleway.com/cockpit/plans) in the Scaleway console.

Refer to Cockpit's [product documentation](https://www.scaleway.com/en/docs/observability/cockpit/concepts/) and [API documentation](https://www.scaleway.com/en/developers/api/cockpit/regional-api) for more information.

## Fetch and associate a pricing plan to a Cockpit

The following command shows how to fetch information about the `premium` pricing plan and how to associate it with the Cockpit of your Scaleway default Project.

```hcl
data "scaleway_cockpit_plan" "premium" {
  name = "premium"
}

resource "scaleway_cockpit" "main" {
  plan = data.scaleway_cockpit_plan.premium.id
}
```

## Argument reference

This section lists the arguments that you can provide to the `scaleway_cockpit_plan` data source to filter and retrieve the desired plan.

- `name` - (Required) Name of the pricing plan you want to retrieve information about.
