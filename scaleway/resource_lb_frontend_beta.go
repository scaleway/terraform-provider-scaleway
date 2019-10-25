package scaleway

import (
	"math"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayLbFrontendBeta() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayLbFrontendBetaCreate,
		Read:   resourceScalewayLbFrontendBetaRead,
		Update: resourceScalewayLbFrontendBetaUpdate,
		Delete: resourceScalewayLbFrontendBetaDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"lb_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "The load-balancer id",
			},
			"backend_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "The load-balancer backend id",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the backend",
			},
			"inbound_port": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, math.MaxUint16),
				Description:  "TCP port to listen on the front side.",
			},
			"timeout_client": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: difSuppressFuncDuration,
				ValidateFunc:     validateDuration(),
				Description:      "Set the maximum inactivity time on the client side",
			},
			"certificate_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "Certificate ID.",
			},
		},
	}
}

func resourceScalewayLbFrontendBetaCreate(d *schema.ResourceData, m interface{}) error {
	lbAPI := getLbAPI(m)

	region, LbID, err := parseRegionalID(d.Get("lb_id").(string))
	if err != nil {
		return err
	}

	name, ok := d.GetOk("name")
	if !ok {
		name = getRandomName("lb-frt")
	}

	backendID := expandID(d.Get("backend_id"))
	certificateID := scw.StringPtr(expandID(d.Get("certificate_id")))
	if *certificateID == "" {
		certificateID = nil
	}

	var createReq = &lb.CreateFrontendRequest{
		Region:        region,
		LbID:          LbID,
		Name:          name.(string),
		InboundPort:   int32(d.Get("inbound_port").(int)),
		BackendID:     backendID,
		TimeoutClient: expandDuration(d.Get("timeout_client")),
		CertificateID: certificateID,
	}
	res, err := lbAPI.CreateFrontend(createReq)
	if err != nil {
		return err
	}

	d.SetId(newRegionalId(region, res.ID))

	return resourceScalewayLbFrontendBetaRead(d, m)
}

func resourceScalewayLbFrontendBetaRead(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := getLbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	res, err := lbAPI.GetFrontend(&lb.GetFrontendRequest{
		Region:     region,
		FrontendID: ID,
	})

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("lb_id", newRegionalId(region, res.Lb.ID))
	d.Set("backend_id", newRegionalId(region, res.Backend.ID))
	d.Set("name", res.Name)
	d.Set("inbound_port", int(res.InboundPort))
	d.Set("timeout_client", flattenDuration(res.TimeoutClient))

	if res.Certificate != nil {
		d.Set("certificate_id", newRegionalId(region, res.Certificate.ID))
	} else {
		d.Set("certificate_id", "")
	}

	return nil
}

func resourceScalewayLbFrontendBetaUpdate(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := getLbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	backendID := expandID(d.Get("backend_id"))
	certificateID := scw.StringPtr(expandID(d.Get("certificate_id")))
	if *certificateID == "" {
		certificateID = nil
	}

	req := &lb.UpdateFrontendRequest{
		Region:        region,
		FrontendID:    ID,
		Name:          d.Get("name").(string),
		InboundPort:   int32(d.Get("inbound_port").(int)),
		BackendID:     backendID,
		TimeoutClient: expandDuration(d.Get("timeout_client")),
		CertificateID: certificateID,
	}

	_, err = lbAPI.UpdateFrontend(req)
	if err != nil {
		return err
	}

	return resourceScalewayLbFrontendBetaRead(d, m)
}

func resourceScalewayLbFrontendBetaDelete(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := getLbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	err = lbAPI.DeleteFrontend(&lb.DeleteFrontendRequest{
		Region:     region,
		FrontendID: ID,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	if is404Error(err) {
		return nil
	}

	return err
}
