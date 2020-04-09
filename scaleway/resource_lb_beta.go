package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
			"keep_ip_on_delete": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"ip_id": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
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

	createReq := &lb.CreateLbRequest{
		Region:         region,
		OrganizationID: d.Get("organization_id").(string),
		Name:           expandOrGenerateString(d.Get("name"), "lb"),
		Type:           d.Get("type").(string),
	}

	if ipID, ok := d.GetOk("ip_id"); ok {
		createReq.IPID = scw.StringPtr(expandID(ipID.(string)))
		_ = d.Set("keep_ip_on_delete", true)
	} else {
		_ = d.Set("keep_ip_on_delete", false)
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
		Timeout: InstanceServerWaitForTimeout,
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

	_ = d.Set("name", res.Name)
	_ = d.Set("region", string(region))
	_ = d.Set("organization_id", res.OrganizationID)
	_ = d.Set("tags", res.Tags)
	_ = d.Set("type", res.Type)
	_ = d.Set("ip_id", newRegionalId(region, res.IP[0].ID))
	_ = d.Set("ip_address", res.IP[0].IPAddress)

	return nil
}

func resourceScalewayLbBetaUpdate(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := lbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("name") || d.HasChange("tags") {

		req := &lb.UpdateLbRequest{
			Region: region,
			LbID:   ID,
			Name:   d.Get("name").(string),
			Tags:   expandStrings(d.Get("tags")),
		}

		_, err = lbAPI.UpdateLb(req)
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

	err = lbAPI.DeleteLb(&lb.DeleteLbRequest{
		Region: region,
		LbID:   ID,
		// This parameter will probably be breaking change when ip pre reservation will exist.
		ReleaseIP: !d.Get("keep_ip_on_delete").(bool),
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
