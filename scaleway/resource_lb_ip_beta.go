package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	lb "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayLbIPBeta() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayLbIPBetaCreate,
		Read:   resourceScalewayLbIPBetaRead,
		Update: resourceScalewayLbIPBetaUpdate,
		Delete: resourceScalewayLbIPBetaDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"reverse": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The reverse domain name for this IP",
			},
			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
			// Computed
			"ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The load-balance public IP address",
			},
			"lb_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the loadbalancer attached to this IP, if any",
			},
		},
	}
}

func resourceScalewayLbIPBetaCreate(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, err := lbAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	createReq := &lb.CreateIPRequest{
		Region:         region,
		OrganizationID: d.Get("organization_id").(string),
	}
	if reverse, ok := d.GetOk("reverse"); ok {
		createReq.Reverse = scw.StringPtr(reverse.(string))
	}
	res, err := lbAPI.CreateIP(createReq)
	if err != nil {
		return err
	}

	d.SetId(newRegionalId(region, res.ID))

	return resourceScalewayLbIPBetaRead(d, m)
}

func resourceScalewayLbIPBetaRead(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := lbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	res, err := lbAPI.GetIP(&lb.GetIPRequest{
		Region: region,
		IPID:   ID,
	})

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	_ = d.Set("region", string(region))
	_ = d.Set("organization_id", res.OrganizationID)
	_ = d.Set("ip_id", res.ID)
	_ = d.Set("ip_address", res.IPAddress)
	_ = d.Set("reverse", res.Reverse)
	if res.LbID != nil {
		_ = d.Set("lb_id", *res.LbID)
	} else {
		_ = d.Set("lb_id", "")
	}

	return nil
}

func resourceScalewayLbIPBetaUpdate(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := lbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("reverse") {
		req := &lb.UpdateIPRequest{
			Region:  region,
			IPID:    ID,
			Reverse: scw.StringPtr(d.Get("reverse").(string)),
		}

		_, err = lbAPI.UpdateIP(req)
		if err != nil {
			return err
		}
	}

	return resourceScalewayLbIPBetaRead(d, m)
}

func resourceScalewayLbIPBetaDelete(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := lbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	err = lbAPI.ReleaseIP(&lb.ReleaseIPRequest{
		Region: region,
		IPID:   ID,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
