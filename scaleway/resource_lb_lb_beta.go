package scaleway

import (
	"github.com/hashicorp/terraform/helper/schema"
	lb "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
)

func resourceScalewayLbLbBeta() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayLbLbBetaCreate,
		Read:   resourceScalewayLbLbBetaRead,
		Update: resourceScalewayLbLbBetaUpdate,
		Delete: resourceScalewayLbLbBetaDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Timeouts: &schema.ResourceTimeout{
			Create: &BaremetalServerResourceTimeout,
			Delete: &BaremetalServerResourceTimeout,
		},
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
			"ips": {
				Type: schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
				Computed:    true,
				Description: "Array of ip ids attached to the load-balancer",
			},
			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
		},
	}
}

func resourceScalewayLbLbBetaCreate(d *schema.ResourceData, m interface{}) error {
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

	return resourceScalewayLbLbBetaRead(d, m)
}

func resourceScalewayLbLbBetaRead(d *schema.ResourceData, m interface{}) error {
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

	ips := []map[string]interface{}(nil)
	for _, ip := range res.IP {
		ips = append(ips, map[string]interface{}{
			"ip_id":   ip.ID,
			"address": ip.IPAddress,
		})
	}

	d.Set("name", res.Name)
	d.Set("region", string(region))
	d.Set("organization_id", res.OrganizationID)
	d.Set("tags", res.Tags)
	d.Set("type", res.Type)
	d.Set("ips", ips)

	return nil
}

func resourceScalewayLbLbBetaUpdate(d *schema.ResourceData, m interface{}) error {
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

	return resourceScalewayLbLbBetaRead(d, m)
}

func resourceScalewayLbLbBetaDelete(d *schema.ResourceData, m interface{}) error {
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

	if is404Error(err) {
		return nil
	}

	return err
}
