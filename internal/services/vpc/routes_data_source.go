package vpc

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func DataSourceRoutes() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceRoutesRead,
		SchemaFunc:  routesSchema,
	}
}

func routesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"vpc_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Only routes within this VPC will be returned",
		},
		"nexthop_resource_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Only routes with a matching next hop resource ID will be returned",
		},
		"nexthop_private_network_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Only routes with a matching next hop private network ID will be returned",
		},
		"nexthop_resource_type": {
			Type:             schema.TypeString,
			Optional:         true,
			Description:      "Only Routes with a matching next hop resource type will be returned",
			ValidateDiagFunc: verify.ValidateEnum[vpc.RouteWithNexthopResourceType](),
		},
		"is_ipv6": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Only routes with an IPv6 destination will be returned",
		},
		"tags": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "Routes with these exact tags are listed.",
		},
		"routes": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of routes",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "The associated route ID",
					},
					"vpc_id": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "The VPC ID associated with the route",
					},
					"tags": {
						Computed: true,
						Type:     schema.TypeList,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
						Description: "The tags associated with the route",
					},
					"created_at": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "Date and time of route's creation (RFC 3339 format)",
					},
					"description": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "The description of the route",
					},
					"destination": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "The destination IP or IP range of the route",
					},
					"nexthop_resource_id": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "The resource ID of the route's next hop",
					},
					"nexthop_private_network_id": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "The private network ID of the route's next hop",
					},
					"nexthop_ip": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "The IP of the route's next hop",
					},
					"nexthop_name": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "The name of the route's next hop",
					},
					"nexthop_resource_type": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "The resource type of the route's next hop",
					},
					"region": regional.Schema(),
				},
			},
		},
		"region": regional.Schema(),
	}
}

func DataSourceRoutesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
		NexthopResourceType:     vpc.RouteWithNexthopResourceType(d.Get("nexthop_resource_type").(string)),
	}

	isipv6, isipv6Exists := d.GetOk("is_ipv6")
	if isipv6Exists {
		req.IsIPv6 = types.ExpandBoolPtr(isipv6)
	}

	res, err := routesAPI.ListRoutesWithNexthop(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	routes := []any(nil)

	for _, route := range res.Routes {
		rawRoute := make(map[string]any)
		if route.Route != nil {
			rawRoute["id"] = regional.NewIDString(region, route.Route.ID)
			rawRoute["created_at"] = types.FlattenTime(route.Route.CreatedAt)
			rawRoute["vpc_id"] = route.Route.VpcID
			rawRoute["nexthop_resource_id"] = types.FlattenStringPtr(route.Route.NexthopResourceID)
			rawRoute["nexthop_private_network_id"] = types.FlattenStringPtr(route.Route.NexthopPrivateNetworkID)
			rawRoute["description"] = route.Route.Description
			rawRoute["region"] = region.String()

			destination, err := types.FlattenIPNet(route.Route.Destination)
			if err != nil {
				return diag.FromErr(err)
			}

			rawRoute["destination"] = destination

			if len(route.Route.Tags) > 0 {
				rawRoute["tags"] = route.Route.Tags
			}
		}

		rawRoute["nexthop_ip"] = types.FlattenIPPtr(route.NexthopIP)
		rawRoute["nexthop_name"] = types.FlattenStringPtr(route.NexthopName)
		rawRoute["nexthop_resource_type"] = route.NexthopResourceType.String()

		routes = append(routes, rawRoute)
	}

	d.SetId(region.String())
	_ = d.Set("routes", routes)

	return nil
}
