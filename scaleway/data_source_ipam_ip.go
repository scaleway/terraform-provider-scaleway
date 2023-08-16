package scaleway

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	ipam "github.com/scaleway/scaleway-sdk-go/api/ipam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func dataSourceScalewayIPAMIP() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayIPAMIPRead,
		Schema: map[string]*schema.Schema{
			// Input
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Servers with these exact tags are listed.",
			},
			"private_network_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"resource_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"mac_address": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "IP Type (ipv4, ipv6)",
			},
			"zone":            zoneSchema(),
			"organization_id": organizationIDSchema(),

			// Computed
			"address": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceScalewayIPAMIPRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api, region, err := ipamAPIWithRegion(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	req := &ipam.ListIPsRequest{
		Region:           region,
		OrganizationID:   nil,
		PrivateNetworkID: expandStringPtr(d.Get("private_network_id")),
		SubnetID:         nil,
		Attached:         nil,
		ResourceID:       expandStringPtr(d.Get("resource_id")),
		ResourceType:     "",
		MacAddress:       expandStringPtr(d.Get("mac_address")),
		Tags:             nil,
		IsIPv6:           nil,
		ResourceName:     nil,
		ResourceIDs:      nil,
	}

	switch d.Get("type").(string) {
	case "ipv4":
		req.IsIPv6 = scw.BoolPtr(false)
	case "ipv6":
		req.IsIPv6 = scw.BoolPtr(true)
	default:
		return diag.Diagnostics{{
			Severity:      diag.Error,
			Summary:       "Invalid IP Type",
			Detail:        "Expected ipv4 or ipv6",
			AttributePath: cty.GetAttrPath("type"),
		}}
	}

	resp, err := api.ListIPs(req, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}
	if len(resp.IPs) == 0 {
		return diag.FromErr(fmt.Errorf("no ip found with given filters"))
	}
	if len(resp.IPs) > 1 {
		return diag.FromErr(fmt.Errorf("more than one ip found with given filter"))
	}

	ip := resp.IPs[0]

	d.SetId(ip.ID)
	_ = d.Set("address", ip.Address.String())

	return nil
}
