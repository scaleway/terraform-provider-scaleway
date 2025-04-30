---
subcategory: "Cockpit"
page_title: "Scaleway: scaleway_cockpit_plan"
---
# scaleway_cockpit_plan

**Note:** As of January 1st, 2025, Cockpit pricing plans have been deprecated. While this data source remains available temporarily for backward compatibility, Scaleway no longer supports configuring Cockpit resources using fixed pricing plans. Instead, you should now independently configure retention periods for your data sources (metrics, logs, and traces). Refer to [Scaleway Cockpit Documentation](https://www.scaleway.com/en/docs/cockpit/concepts/#retention) for updated guidelines and [pricing information](https://www.scaleway.com/en/docs/cockpit/faq/#how-am-i-billed-for-increasing-data-retention-period).

The `scaleway_cockpit_plan` data source retrieves details about a specific Scaleway Cockpit pricing plan. You can use this data source to manage existing Terraform configurations that reference Cockpit plans.

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
