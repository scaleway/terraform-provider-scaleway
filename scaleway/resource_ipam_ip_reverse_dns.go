package scaleway

import (
	"context"
	"net"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
)

func resourceScalewayIPAMIPReverseDNS() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayIPAMIPReverseDNSCreate,
		ReadContext:   resourceScalewayIPAMIPReverseDNSRead,
		UpdateContext: resourceScalewayIPAMIPReverseDNSUpdate,
		DeleteContext: resourceScalewayIPAMIPReverseDNSDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultIPAMIPReverseDNSTimeout),
			Create:  schema.DefaultTimeout(defaultIPAMIPReverseDNSTimeout),
			Update:  schema.DefaultTimeout(defaultIPAMIPReverseDNSTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"ipam_ip_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The IPAM IP ID",
			},
			"hostname": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The reverse domain name",
			},
			"address": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The IP corresponding to the hostname",
				ValidateFunc: validation.IsIPAddress,
			},
			"region": regional.Schema(),
		},
	}
}

func resourceScalewayIPAMIPReverseDNSCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ipamAPI, region, err := ipamAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := ipamAPI.GetIP(&ipam.GetIPRequest{
		Region: region,
		IPID:   locality.ExpandID(d.Get("ipam_ip_id")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(regional.NewIDString(region, res.ID))
	if hostname, ok := d.GetOk("hostname"); ok {
		reverse := &ipam.Reverse{
			Hostname: hostname.(string),
			Address:  scw.IPPtr(net.ParseIP(d.Get("address").(string))),
		}

		updateReverseReq := &ipam.UpdateIPRequest{
			Region:   region,
			IPID:     res.ID,
			Reverses: []*ipam.Reverse{reverse},
		}

		_, err := ipamAPI.UpdateIP(updateReverseReq, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayIPAMIPReverseDNSRead(ctx, d, m)
}

func resourceScalewayIPAMIPReverseDNSRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ipamAPI, region, ID, err := ipamAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := ipamAPI.GetIP(&ipam.GetIPRequest{
		Region: region,
		IPID:   ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	managedHostname := d.Get("hostname").(string)
	managedAddress := d.Get("address").(string)
	for _, reverse := range res.Reverses {
		if reverse.Hostname == managedHostname && reverse.Address.String() == managedAddress {
			_ = d.Set("hostname", reverse.Hostname)
			_ = d.Set("address", flattenIPPtr(reverse.Address))
			break
		}
	}

	_ = d.Set("region", region)

	return nil
}

func resourceScalewayIPAMIPReverseDNSUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ipamAPI, region, ID, err := ipamAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChanges("hostname", "address") {
		reverse := &ipam.Reverse{
			Hostname: d.Get("hostname").(string),
			Address:  scw.IPPtr(net.ParseIP(d.Get("address").(string))),
		}

		updateReverseReq := &ipam.UpdateIPRequest{
			Region:   region,
			IPID:     ID,
			Reverses: []*ipam.Reverse{reverse},
		}

		_, err := ipamAPI.UpdateIP(updateReverseReq, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayIPAMIPReverseDNSRead(ctx, d, m)
}

func resourceScalewayIPAMIPReverseDNSDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	ipamAPI, region, ID, err := ipamAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	updateReverseReq := &ipam.UpdateIPRequest{
		Region:   region,
		IPID:     ID,
		Reverses: []*ipam.Reverse{},
	}

	_, err = ipamAPI.UpdateIP(updateReverseReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
