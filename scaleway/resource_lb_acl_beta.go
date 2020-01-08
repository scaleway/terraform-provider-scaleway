package scaleway

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	lb "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
)

func resourceScalewayLbAclBeta() *schema.Resource {
	return &schema.Resource{
		Create: resourceScalewayLbAclBetaCreate,
		Read:   resourceScalewayLbAclBetaRead,
		Update: resourceScalewayLbAclBetaUpdate,
		Delete: resourceScalewayLbAclBetaDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of you ACL resource.",
			},
			"frontend_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "ID of your frontend.",
			},
			"action": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Action to undertake",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								lb.ACLActionTypeAllow.String(),
								lb.ACLActionTypeDeny.String(),
							}, false),
							Description: "<allow> or <deny> request",
						},
					},
				},
			},
			"match": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "AclMatch Rule",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_subnet": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional: true,
							Description: "This is the source IP v4/v6 address of the client of the session to match or not. " +
								"Addresses values can be specified either as plain addresses or with a netmask appended.",
						},
						"http_filter": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								lb.ACLHTTPFilterACLHTTPFilterNone.String(),
								lb.ACLHTTPFilterPathBegin.String(),
								lb.ACLHTTPFilterPathEnd.String(),
								lb.ACLHTTPFilterRegex.String(),
							}, false),
							Description: "Http filter (if backend have a http forward protocol)",
						},
						"http_filter_value": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Http filter value",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"invert": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "If true, then condition is unless type",
						},
					},
				},
			},
			"index": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Order between your ACLs (ascending order, 0 is first acl executed).",
			},
			"region":          regionSchema(),
			"organization_id": organizationIDSchema(),
		},
	}
}

func resourceScalewayLbAclBetaCreate(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, err := getLbAPIWithRegion(d, m)
	if err != nil {
		return err
	}

	createReq := &lb.CreateACLRequest{
		Region:     region,
		FrontendID: expandID(d.Get("frontend_id")),
		Name:       expandOrGenerateString(d.Get("name"), "lb-acl"),
		Action:     expandAclAction(d.Get("action")),
		Match:      expandAclMatch(d.Get("match")),
		Index:      int32(d.Get("index").(int)),
	}

	res, err := lbAPI.CreateACL(createReq)
	if err != nil {
		return err
	}

	d.SetId(newRegionalId(region, res.ID))

	return resourceScalewayLbAclBetaRead(d, m)
}

func resourceScalewayLbAclBetaRead(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := getLbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	res, err := lbAPI.GetACL(&lb.GetACLRequest{
		Region: region,
		ACLID:  ID,
	})

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	_ = d.Set("name", res.Name)
	_ = d.Set("frontend_id", newRegionalId(region, res.Frontend.ID))
	_ = d.Set("action", res.Action)
	_ = d.Set("match", res.Match)
	_ = d.Set("index", res.Index)

	return nil
}

func resourceScalewayLbAclBetaUpdate(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := getLbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	req := &lb.UpdateACLRequest{
		Region: region,
		ACLID:  ID,
		Name:   d.Get("name").(string),
		Action: expandAclAction(d.Get("action")),
		Match:  expandAclMatch(d.Get("match")),
		Index:  int32(d.Get("index").(int)),
	}

	_, err = lbAPI.UpdateACL(req)
	if err != nil {
		return err
	}

	return resourceScalewayLbAclBetaRead(d, m)
}

func resourceScalewayLbAclBetaDelete(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := getLbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	err = lbAPI.DeleteACL(&lb.DeleteACLRequest{
		Region: region,
		ACLID:  ID,
	})

	if err != nil && !is404Error(err) {
		return err
	}

	return nil
}
