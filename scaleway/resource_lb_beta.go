package scaleway

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	lb "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayLbBeta() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayLbBetaCreate,
		Read:   resourceScalewayLbBetaRead,
		Update: resourceScalewayLbBetaUpdate,
		Delete: resourceScalewayLbBetaDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the lb",
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: diffSuppressFuncIgnoreCase,
				Description:      "The type of load-balancer you want to create",
			},
			"tags": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Array of tags to associate with the load-balancer",
			},
			"ip_id": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "The load-balance public IP ID",
				ForceNew:         true,
				DiffSuppressFunc: diffSuppressFuncLocality,
			},
			"ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The load-balance public IP address",
			},
			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
		},
	}
}

func resourceScalewayLbBetaCreate(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, err := lbAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	createReq := &lb.CreateLBRequest{
		Region:         region,
		IPID:           scw.StringPtr(expandID(d.Get("ip_id").(string))),
		OrganizationID: expandStringPtr(d.Get("organization_id")),
		Name:           expandOrGenerateString(d.Get("name"), "lb"),
		Type:           d.Get("type").(string),
	}

	if raw, ok := d.GetOk("tags"); ok {
		for _, tag := range raw.([]interface{}) {
			createReq.Tags = append(createReq.Tags, tag.(string))
		}
	}
	res, err := lbAPI.CreateLB(createReq)
	if err != nil {
		return err
	}

	d.SetId(newRegionalIDString(region, res.ID))

	_, err = lbAPI.WaitForLb(&lb.WaitForLBRequest{
		Region:  region,
		LBID:    res.ID,
		Timeout: scw.TimeDurationPtr(InstanceServerWaitForTimeout),
	})
	if err != nil {
		return err
	}

	return resourceScalewayLbBetaRead(d, m)
}

func resourceScalewayLbBetaRead(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := lbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	res, err := lbAPI.GetLB(&lb.GetLBRequest{
		Region: region,
		LBID:   ID,
	})

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	_ = d.Set("name", res.Name)
	_ = d.Set("region", string(region))
	_ = d.Set("organization_id", res.OrganizationID)
	_ = d.Set("tags", res.Tags)
	// For now API return lowercase lb type. This should be fix in a near future on the API side
	_ = d.Set("type", strings.ToUpper(res.Type))
	_ = d.Set("ip_id", newRegionalIDString(region, res.IP[0].ID))
	_ = d.Set("ip_address", res.IP[0].IPAddress)

	return nil
}

func resourceScalewayLbBetaUpdate(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := lbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("name") || d.HasChange("tags") {
		req := &lb.UpdateLBRequest{
			Region: region,
			LBID:   ID,
			Name:   d.Get("name").(string),
			Tags:   expandStrings(d.Get("tags")),
		}

		_, err = lbAPI.UpdateLB(req)
		if err != nil {
			return err
		}
	}

	return resourceScalewayLbBetaRead(d, m)
}

func resourceScalewayLbBetaDelete(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := lbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	err = lbAPI.DeleteLB(&lb.DeleteLBRequest{
		Region:    region,
		LBID:      ID,
		ReleaseIP: false,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	_, err = lbAPI.WaitForLb(&lb.WaitForLBRequest{
		LBID:    ID,
		Region:  region,
		Timeout: scw.TimeDurationPtr(LbWaitForTimeout),
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
