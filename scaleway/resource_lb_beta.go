package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	lb "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
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
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The type of load-balancer you want to create",
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
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The load-balance public IP ID",
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
	lbAPI, region, err := getLbAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	name, ok := d.GetOk("name")
	if !ok {
		name = getRandomName("lb")
	}
	createReq := &lb.CreateLbRequest{
		Region:         region,
		OrganizationID: d.Get("organization_id").(string),
		Name:           name.(string),
		Type:           d.Get("type").(string),
	}
	if raw, ok := d.GetOk("tags"); ok {
		for _, tag := range raw.([]interface{}) {
			createReq.Tags = append(createReq.Tags, tag.(string))
		}
	}
	res, err := lbAPI.CreateLb(createReq)
	if err != nil {
		return err
	}

	d.SetId(newRegionalId(region, res.ID))

	_, err = lbAPI.WaitForLb(&lb.WaitForLbRequest{
		Region:  region,
		LbID:    res.ID,
		Timeout: BaremetalServerWaitForTimeout,
	})
	if err != nil {
		return err
	}

	return resourceScalewayLbBetaRead(d, m)
}

func resourceScalewayLbBetaRead(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := getLbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	res, err := lbAPI.GetLb(&lb.GetLbRequest{
		Region: region,
		LbID:   ID,
	})

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("name", res.Name)
	d.Set("region", string(region))
	d.Set("organization_id", res.OrganizationID)
	d.Set("tags", res.Tags)
	d.Set("type", res.Type)
	d.Set("ip_id", res.IP[0].ID)
	d.Set("ip_address", res.IP[0].IPAddress)

	return nil
}

func resourceScalewayLbBetaUpdate(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := getLbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("name") || d.HasChange("tags") {

		req := &lb.UpdateLbRequest{
			Region: region,
			LbID:   ID,
			Name:   d.Get("name").(string),
			Tags:   StringSliceFromState(d.Get("tags").([]interface{})),
		}

		_, err = lbAPI.UpdateLb(req)
		if err != nil {
			return err
		}
	}

	return resourceScalewayLbBetaRead(d, m)
}

func resourceScalewayLbBetaDelete(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := getLbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	err = lbAPI.DeleteLb(&lb.DeleteLbRequest{
		Region: region,
		LbID:   ID,
		// This parameter will probably be breaking change when ip pre reservation will exist.
		ReleaseIP: true,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	_, err = lbAPI.WaitForLb(&lb.WaitForLbRequest{
		LbID:    ID,
		Region:  region,
		Timeout: LbWaitForTimeout,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
