package scaleway

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
)

func resourceScalewayLbCertificateBeta() *schema.Resource {
	return &schema.Resource{
		Create:        resourceScalewayLbCertificateBetaCreate,
		Read:          resourceScalewayLbCertificateBetaRead,
		Update:        resourceScalewayLbCertificateBetaUpdate,
		Delete:        resourceScalewayLbCertificateBetaDelete,
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"lb_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The load-balancer ID",
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the load-balancer certificate",
				Optional:    true,
				Computed:    true,
			},
			"letsencrypt": {
				ConflictsWith: []string{"custom_certificate"},
				MaxItems:      1,
				Description:   "The Let's Encrypt type certificate configuration",
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"common_name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The main domain name of the certificate",
						},
						"subject_alternative_name": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional:    true,
							Description: "The alternative domain names of the certificate",
						},
					},
				},
			},
			"custom_certificate": {
				ConflictsWith: []string{"letsencrypt"},
				MaxItems:      1,
				Type:          schema.TypeList,
				Description:   "The custom type certificate type configuration",
				Optional:      true,
				ForceNew:      true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_chain": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The full PEM-formatted certificate chain",
						},
					},
				},
			},

			// Readonly attributes
			"common_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The main domain name of the certificate",
			},
			"subject_alternative_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The alternative domain names of the certificate",
			},
			"fingerprint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The identifier (SHA-1) of the certificate",
			},
			"not_valid_before": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The not valid before validity bound timestamp",
			},
			"not_valid_after": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The not valid after validity bound timestamp",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The status of certificate",
			},
		},
	}
}

func resourceScalewayLbCertificateBetaCreate(d *schema.ResourceData, m interface{}) error {
	region, lbID, err := parseRegionalID(d.Get("lb_id").(string))
	if err != nil {
		return err
	}

	createReq := &lb.CreateCertificateRequest{
		Region:            region,
		LBID:              lbID,
		Name:              expandOrGenerateString(d.Get("name"), "lb-cert"),
		Letsencrypt:       expandLbLetsEncrypt(d.Get("letsencrypt")),
		CustomCertificate: expandLbCustomCertificate(d.Get("custom_certificate")),
	}
	if createReq.Letsencrypt == nil && createReq.CustomCertificate == nil {
		return errors.New("you need to define either letsencrypt or custom_certificate configuration")
	}

	lbAPI := lbAPI(m)
	res, err := lbAPI.CreateCertificate(createReq)
	if err != nil {
		return err
	}

	d.SetId(newRegionalId(region, res.ID))

	return resourceScalewayLbCertificateBetaRead(d, m)
}

func resourceScalewayLbCertificateBetaRead(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := lbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	res, err := lbAPI.GetCertificate(&lb.GetCertificateRequest{
		CertificateID: ID,
		Region:        region,
	})

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	_ = d.Set("name", res.Name)
	_ = d.Set("common_name", res.CommonName)
	_ = d.Set("subject_alternative_name", res.SubjectAlternativeName)
	_ = d.Set("fingerprint", res.Fingerprint)
	_ = d.Set("not_valid_before", flattenTime(res.NotValidBefore))
	_ = d.Set("not_valid_after", flattenTime(res.NotValidAfter))
	_ = d.Set("status", res.Status)
	return nil
}

func resourceScalewayLbCertificateBetaUpdate(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := lbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	req := &lb.UpdateCertificateRequest{
		CertificateID: ID,
		Region:        region,
		Name:          d.Get("name").(string),
	}

	_, err = lbAPI.UpdateCertificate(req)
	if err != nil {
		return err
	}

	return resourceScalewayLbCertificateBetaRead(d, m)
}

func resourceScalewayLbCertificateBetaDelete(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := lbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	err = lbAPI.DeleteCertificate(&lb.DeleteCertificateRequest{
		Region:        region,
		CertificateID: ID,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
