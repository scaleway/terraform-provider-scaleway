package instance

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceServers() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceInstanceServersRead,
		Schema: map[string]*schema.Schema{
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
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"public_ip": {
							Computed:   true,
							Type:       schema.TypeString,
							Deprecated: "Use public_ips instead",
						},
						"public_ips": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"address": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"private_ip": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"state": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"name": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"boot_type": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"bootscript_id": {
							Computed:   true,
							Type:       schema.TypeString,
							Deprecated: "bootscript are not supported",
						},
						"type": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"tags": {
							Computed: true,
							Type:     schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"security_group_id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"enable_ipv6": {
							Computed: true,
							Type:     schema.TypeBool,
						},
						"enable_dynamic_ip": {
							Computed: true,
							Type:     schema.TypeBool,
						},
						"image": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"placement_group_id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"placement_group_policy_respected": {
							Computed: true,
							Type:     schema.TypeBool,
						},
						"ipv6_address": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"ipv6_gateway": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"ipv6_prefix_length": {
							Computed: true,
							Type:     schema.TypeInt,
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
		},
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

		if server.PublicIP != nil { //nolint:staticcheck
			rawServer["public_ip"] = server.PublicIP.Address.String() //nolint:staticcheck
		}

		if server.PublicIPs != nil {
			rawServer["public_ips"] = flattenServerPublicIPs(server.Zone, server.PublicIPs)
		}

		if server.PrivateIP != nil {
			rawServer["private_ip"] = *server.PrivateIP
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
		if server.EnableIPv6 != nil { //nolint:staticcheck
			rawServer["enable_ipv6"] = server.EnableIPv6 //nolint:staticcheck
		}

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

		if server.IPv6 != nil { //nolint:staticcheck
			rawServer["ipv6_address"] = server.IPv6.Address.String() //nolint:staticcheck
			rawServer["ipv6_gateway"] = server.IPv6.Gateway.String() //nolint:staticcheck

			prefixLength, err := strconv.Atoi(server.IPv6.Netmask) //nolint:staticcheck
			if err != nil {
				diags = append(diags, diag.FromErr(fmt.Errorf("failed to read ipv6 netmask: %w", err))...)

				continue
			}

			rawServer["ipv6_prefix_length"] = prefixLength
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
