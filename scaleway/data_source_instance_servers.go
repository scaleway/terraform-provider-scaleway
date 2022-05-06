package scaleway

import (
	"context"

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

	servers := []interface{}(nil)
	for _, server := range res.Servers {
		rawServer := make(map[string]interface{})
		rawServer["id"] = newZonedID(server.Zone, server.ID).String()
		if server.PublicIP != nil {
			rawServer["public_ip"] = server.PublicIP.Address.String()
		}
		servers = append(servers, rawServer)
	}

	d.SetId(zone.String())
	_ = d.Set("servers", servers)

	return nil
}
