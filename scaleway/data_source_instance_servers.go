package scaleway

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayInstanceServers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayInstanceServersRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Servers with a name matching it are listed.",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Servers with these exact tags are listed. Use commas to separate multiple tags when filtering.",
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
							Computed: true,
							Type:     schema.TypeString,
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
							Computed: true,
							Type:     schema.TypeString,
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
						"zone":            zoneSchema(),
						"organization_id": organizationIDSchema(),
						"project_id":      projectIDSchema(),
					},
				},
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
		},
	}
}

func dataSourceScalewayInstanceServersRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceAPI, zone, err := instanceAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	res, err := instanceAPI.ListServers(&instance.ListServersRequest{
		Zone:    zone,
		Name:    expandStringPtr(d.Get("name")),
		Project: expandStringPtr(d.Get("project_id")),
		Tags:    expandStrings(d.Get("tags")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	var diags diag.Diagnostics

	servers := []interface{}(nil)
	for _, server := range res.Servers {
		rawServer := make(map[string]interface{})
		rawServer["id"] = newZonedID(server.Zone, server.ID).String()
		if server.PublicIP != nil {
			rawServer["public_ip"] = server.PublicIP.Address.String()
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
		rawServer["bootscript_id"] = server.Bootscript.ID
		rawServer["type"] = server.CommercialType
		if len(server.Tags) > 0 {
			rawServer["tags"] = server.Tags
		}
		rawServer["security_group_id"] = newZonedID(zone, server.SecurityGroup.ID).String()
		rawServer["enable_ipv6"] = server.EnableIPv6
		rawServer["enable_dynamic_ip"] = server.DynamicIPRequired
		rawServer["organization_id"] = server.Organization
		rawServer["project_id"] = server.Project
		if server.Image != nil {
			rawServer["image"] = server.Image.ID
		}
		if server.PlacementGroup != nil {
			rawServer["placement_group_id"] = newZonedID(zone, server.PlacementGroup.ID).String()
			rawServer["placement_group_policy_respected"] = server.PlacementGroup.PolicyRespected
		}
		if server.IPv6 != nil {
			rawServer["ipv6_address"] = server.IPv6.Address.String()
			rawServer["ipv6_gateway"] = server.IPv6.Gateway.String()
			prefixLength, err := strconv.Atoi(server.IPv6.Netmask)
			if err != nil {
				diags = append(diags, diag.FromErr(fmt.Errorf("failed to read ipv6 netmask: %w", err))...)
				continue
			} else {
				rawServer["ipv6_prefix_length"] = prefixLength
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
