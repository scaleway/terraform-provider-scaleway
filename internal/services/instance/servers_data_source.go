package instance

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	ipamAPI "github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/ipam"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

//go:embed descriptions/servers_datasource.md
var serversDataSourceDescription string

func DataSourceServers() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceInstanceServersRead,
		SchemaFunc:  serversSchema,
		Description: serversDataSourceDescription,
	}
}

func serversSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Servers with a name like it are listed.",
		},
		"tags": {
			Type: schema.TypeList,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Optional:    true,
			Description: "Servers with these exact tags are listed.",
		},
		"servers": {
			Type:        schema.TypeList,
			Description: "Servers",
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"id": {
						Computed:    true,
						Description: "UUID of the server.",
						Type:        schema.TypeString,
					},
					"public_ips": {
						Type:        schema.TypeList,
						Description: "Public IPs associated with this server.",
						Computed:    true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"id": {
									Type:        schema.TypeString,
									Description: "UUID of the public IP.",
									Computed:    true,
								},
								"address": {
									Type:        schema.TypeString,
									Description: "Address of the server",
									Computed:    true,
								},
							},
						},
					},
					"private_ips": {
						Type:        schema.TypeList,
						Computed:    true,
						Optional:    true,
						Description: "List of private IPv4 and IPv6 addresses associated with the resource",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"id": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The ID of the IPv4/v6 address resource",
								},
								"address": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The private IPv4/v6 address",
								},
							},
						},
					},
					"state": {
						Computed:    true,
						Description: "State of the server",
						Type:        schema.TypeString,
					},
					"name": {
						Computed:    true,
						Description: "Name of the server",
						Type:        schema.TypeString,
					},
					"boot_type": {
						Computed:    true,
						Description: "Boot type",
						Type:        schema.TypeString,
					},
					"bootscript_id": {
						Computed:    true,
						Type:        schema.TypeString,
						Description: "UUID of the bootscript",
						Deprecated:  "bootscript are not supported",
					},
					"type": {
						Computed:    true,
						Description: "Type of the server",
						Type:        schema.TypeString,
					},
					"tags": {
						Computed:    true,
						Description: "List of tags assigned to the server.",
						Type:        schema.TypeList,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
					"security_group_id": {
						Computed:    true,
						Description: "Security group ID",
						Type:        schema.TypeString,
					},
					"enable_dynamic_ip": {
						Computed:    true,
						Description: "Whether to enable dynamic IP addresses on this server",
						Type:        schema.TypeBool,
					},
					"image": {
						Computed:    true,
						Description: "Image ID of the server",
						Type:        schema.TypeString,
					},
					"placement_group_id": {
						Computed:    true,
						Description: "Placement Group ID",
						Type:        schema.TypeString,
					},
					"placement_group_policy_respected": {
						Computed:    true,
						Description: "Whether the placement group policy respected or not",
						Type:        schema.TypeBool,
					},
					"zone":            zonal.Schema(),
					"organization_id": account.OrganizationIDSchema(),
					"project_id":      account.ProjectIDSchema(),
				},
			},
		},
		"zone":            zonal.Schema(),
		"organization_id": account.OrganizationIDSchema(),
		"project_id":      account.ProjectIDSchema(),
	}
}

func DataSourceInstanceServersRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	instanceAPI, zone, err := newAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := instanceAPI.ListServers(&instance.ListServersRequest{
		Zone:    zone,
		Name:    types.ExpandStringPtr(d.Get("name")),
		Project: types.ExpandStringPtr(d.Get("project_id")),
		Tags:    types.ExpandStrings(d.Get("tags")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	servers := []any(nil)

	for _, server := range res.Servers {
		rawServer := make(map[string]any)
		rawServer["id"] = zonal.NewID(server.Zone, server.ID).String()

		if server.PublicIPs != nil {
			rawServer["public_ips"] = flattenServerPublicIPs(server.Zone, server.PublicIPs)
		}

		state, err := serverStateFlatten(server.State)
		if err != nil {
			diags = append(diags, diag.FromErr(err)...)

			continue
		}

		rawServer["state"] = state
		rawServer["zone"] = string(zone)
		rawServer["name"] = server.Name
		rawServer["boot_type"] = server.BootType
		rawServer["type"] = server.CommercialType

		if len(server.Tags) > 0 {
			rawServer["tags"] = server.Tags
		}

		rawServer["security_group_id"] = zonal.NewID(zone, server.SecurityGroup.ID).String()

		rawServer["enable_dynamic_ip"] = server.DynamicIPRequired
		rawServer["organization_id"] = server.Organization
		rawServer["project_id"] = server.Project

		if server.Image != nil {
			rawServer["image"] = server.Image.ID
		}

		if server.PlacementGroup != nil {
			rawServer["placement_group_id"] = zonal.NewID(zone, server.PlacementGroup.ID).String()
			rawServer["placement_group_policy_respected"] = server.PlacementGroup.PolicyRespected
		}

		ph, err := newPrivateNICHandler(instanceAPI, server.ID, zone)
		if err != nil {
			return diag.FromErr(err)
		}

		privateNICIDs := []string(nil)
		for _, nic := range ph.privateNICsMap {
			privateNICIDs = append(privateNICIDs, nic.ID)
		}

		// Read server's private IPs if possible
		allPrivateIPs := []map[string]any(nil)
		resourceType := ipamAPI.ResourceTypeInstancePrivateNic

		region, err := zone.Region()
		if err != nil {
			return append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Unable to get server's private IPs",
				Detail:   err.Error(),
			})
		}

		for _, nicID := range privateNICIDs {
			opts := &ipam.GetResourcePrivateIPsOptions{
				ResourceType: &resourceType,
				ResourceID:   &nicID,
				ProjectID:    &server.Project,
			}

			privateIPs, err := ipam.GetResourcePrivateIPs(ctx, m, region, opts)

			switch {
			case err == nil:
				allPrivateIPs = append(allPrivateIPs, privateIPs...)
			case httperrors.Is403(err):
				return append(diags, diag.Diagnostic{
					Severity:      diag.Warning,
					Summary:       "Unauthorized to read server's private IPs, please check your IAM permissions",
					Detail:        err.Error(),
					AttributePath: cty.GetAttrPath("private_ips"),
				})
			default:
				diags = append(diags, diag.Diagnostic{
					Severity:      diag.Warning,
					Summary:       fmt.Sprintf("Unable to get private IPs for server %s (pnic_id: %s)", server.ID, nicID),
					Detail:        err.Error(),
					AttributePath: cty.GetAttrPath("private_ips"),
				})
			}

			if len(allPrivateIPs) > 0 {
				rawServer["private_ips"] = allPrivateIPs
			}
		}

		servers = append(servers, rawServer)
	}

	if len(diags) > 0 {
		return diags
	}

	d.SetId(zone.String())
	_ = d.Set("servers", servers)

	return nil
}
