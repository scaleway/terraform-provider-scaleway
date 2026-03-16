package vpc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/datasource"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceRoute() *schema.Resource {
	dsSchema := datasource.SchemaFromResourceSchema(ResourceRoute().SchemaFunc())

	filterFields := []string{"vpc_id", "nexthop_resource_id", "nexthop_private_network_id", "nexthop_resource_type", "is_ipv6", "tags"}

	dsSchema["route_id"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "The ID of the route",
		ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
		ConflictsWith:    filterFields,
	}

	datasource.AddOptionalFieldsToSchema(dsSchema, "vpc_id", "nexthop_resource_id", "nexthop_private_network_id", "tags", "region")

	for _, key := range []string{"vpc_id", "nexthop_resource_id", "nexthop_private_network_id", "tags"} {
		dsSchema[key].ConflictsWith = []string{"route_id"}
	}

	dsSchema["nexthop_resource_type"] = &schema.Schema{
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "Only routes with a matching next hop resource type will be returned",
		ValidateDiagFunc: verify.ValidateEnum[vpc.RouteWithNexthopResourceType](),
		ConflictsWith:    []string{"route_id"},
	}

	dsSchema["is_ipv6"] = &schema.Schema{
		Type:          schema.TypeBool,
		Optional:      true,
		Description:   "Only routes with an IPv6 destination will be returned",
		ConflictsWith: []string{"route_id"},
	}

	return &schema.Resource{
		ReadContext: DataSourceRouteRead,
		Schema:      dsSchema,
	}
}

func DataSourceRouteRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	routeID, routeIDExists := d.GetOk("route_id")
	if routeIDExists {
		return dataSourceRouteReadByID(ctx, d, m, routeID.(string))
	}

	return dataSourceRouteReadByFilters(ctx, d, m)
}

func dataSourceRouteReadByID(ctx context.Context, d *schema.ResourceData, m any, routeID string) diag.Diagnostics {
	vpcAPI, region, err := vpcAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	regionalID := datasource.NewRegionalID(routeID, region)
	d.SetId(regionalID)
	_ = d.Set("route_id", regionalID)

	res, err := vpcAPI.GetRoute(&vpc.GetRouteRequest{
		Region:  region,
		RouteID: locality.ExpandID(routeID),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setRouteState(d, res)
}

func dataSourceRouteReadByFilters(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	routesAPI, region, err := routesAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &vpc.RoutesWithNexthopAPIListRoutesWithNexthopRequest{
		Region:                  region,
		Tags:                    types.ExpandStrings(d.Get("tags")),
		VpcID:                   types.ExpandStringPtr(locality.ExpandID(d.Get("vpc_id"))),
		NexthopResourceID:       types.ExpandStringPtr(locality.ExpandID(d.Get("nexthop_resource_id"))),
		NexthopPrivateNetworkID: types.ExpandStringPtr(locality.ExpandID(d.Get("nexthop_private_network_id"))),
	}

	nexthopResourceType, nexthopResourceTypeExists := d.GetOk("nexthop_resource_type")
	if nexthopResourceTypeExists {
		req.NexthopResourceType = vpc.RouteWithNexthopResourceType(nexthopResourceType.(string))
	}

	isIPv6, isIPv6Exists := d.GetOk("is_ipv6")
	if isIPv6Exists {
		req.IsIPv6 = types.ExpandBoolPtr(isIPv6)
	}

	res, err := routesAPI.ListRoutesWithNexthop(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	// Filter to only keep custom routes
	var filtered []*vpc.RouteWithNexthop
	for _, r := range res.Routes {
		if r.Route != nil && r.Route.ID != "" {
			filtered = append(filtered, r)
		}
	}

	if len(filtered) == 0 {
		return diag.FromErr(fmt.Errorf("no route found matching the specified filters"))
	}

	if len(filtered) > 1 {
		return diag.FromErr(fmt.Errorf("multiple routes (%d) found, please refine your filters to match exactly one route", len(filtered)))
	}

	route := filtered[0]

	routeRegionalID := regional.NewIDString(route.Route.Region, route.Route.ID)
	d.SetId(routeRegionalID)
	_ = d.Set("route_id", routeRegionalID)

	return setRouteState(d, route.Route)
}
