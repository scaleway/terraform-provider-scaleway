package scaleway

import (
	"math"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				Description:  "The load-balancer ID",
			},
			"backend_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "The load-balancer backend ID",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the frontend",
			},
			"inbound_port": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(0, math.MaxUint16),
				Description:  "TCP port to listen on the front side",
			},
			"timeout_client": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: diffSuppressFuncDuration,
				ValidateFunc:     validateDuration(),
				Description:      "Set the maximum inactivity time on the client side",
			},
			"certificate_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "Certificate ID",
			},
			"acl": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "ACL rules",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The ACL name",
						},
						"action": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "Action to undertake when an ACL filter matches",
							MaxItems:    1,
							MinItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.StringInSlice([]string{
											lb.ACLActionTypeAllow.String(),
											lb.ACLActionTypeDeny.String(),
										}, false),
										Description: "The action type",
									},
								},
							},
						},
						"match": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							MinItems:    1,
							Description: "The ACL match rule",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip_subnet": {
										Type: schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional:    true,
										Description: "A list of IPs or CIDR v4/v6 addresses of the client of the session to match",
									},
									"http_filter": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  lb.ACLHTTPFilterACLHTTPFilterNone.String(),
										ValidateFunc: validation.StringInSlice([]string{
											lb.ACLHTTPFilterACLHTTPFilterNone.String(),
											lb.ACLHTTPFilterPathBegin.String(),
											lb.ACLHTTPFilterPathEnd.String(),
											lb.ACLHTTPFilterRegex.String(),
										}, false),
										Description: "The HTTP filter to match",
									},
									"http_filter_value": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "A list of possible values to match for the given HTTP filter",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"invert": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: `If set to true, the condition will be of type "unless"`,
									},
								},
							},
						},
						"region":          regionSchema(),
						"organization_id": organizationIDSchema(),
					},
				},
			},
		},
	}
}

func resourceScalewayLbFrontendBetaCreate(d *schema.ResourceData, m interface{}) error {
	lbAPI := lbAPI(m)

	region, LbID, err := parseRegionalID(d.Get("lb_id").(string))
	if err != nil {
		return err
	}

	res, err := lbAPI.CreateFrontend(&lb.CreateFrontendRequest{
		Region:        region,
		LBID:          LbID,
		Name:          expandOrGenerateString(d.Get("name"), "lb-frt"),
		InboundPort:   int32(d.Get("inbound_port").(int)),
		BackendID:     expandID(d.Get("backend_id")),
		TimeoutClient: expandDuration(d.Get("timeout_client")),
		CertificateID: expandStringPtr(expandID(d.Get("certificate_id"))),
	})
	if err != nil {
		return err
	}

	d.SetId(newRegionalIDString(region, res.ID))

	err = resourceScalewayLbFrontendBetaUpdateACL(d, lbAPI, region, res.ID)
	if err != nil {
		return err
	}

	return resourceScalewayLbFrontendBetaRead(d, m)
}

func resourceScalewayLbFrontendBetaRead(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := lbAPIWithRegionAndID(m, d.Id())
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

	_ = d.Set("lb_id", newRegionalIDString(region, res.LB.ID))
	_ = d.Set("backend_id", newRegionalIDString(region, res.Backend.ID))
	_ = d.Set("name", res.Name)
	_ = d.Set("inbound_port", int(res.InboundPort))
	_ = d.Set("timeout_client", flattenDuration(res.TimeoutClient))

	if res.Certificate != nil {
		_ = d.Set("certificate_id", newRegionalIDString(region, res.Certificate.ID))
	} else {
		_ = d.Set("certificate_id", "")
	}

	//read related acls.
	resACL, err := lbAPI.ListACLs(&lb.ListACLsRequest{
		Region:     region,
		FrontendID: ID,
	}, scw.WithAllPages())
	if err != nil {
		return err
	}
	sort.Slice(resACL.ACLs, func(i, j int) bool {
		return resACL.ACLs[i].Index < resACL.ACLs[j].Index
	})
	stateAcls := make([]interface{}, 0, len(resACL.ACLs))
	for _, apiACL := range resACL.ACLs {
		stateAcls = append(stateAcls, flattenLbACL(apiACL))
	}
	_ = d.Set("acl", stateAcls)

	return nil
}

func resourceScalewayLbFrontendBetaUpdateACL(d *schema.ResourceData, lbAPI *lb.API, region scw.Region, frontendID string) error {
	//Fetch existing acl from the api. and convert it to a hashmap with index as key
	resACL, err := lbAPI.ListACLs(&lb.ListACLsRequest{
		Region:     region,
		FrontendID: frontendID,
	}, scw.WithAllPages())
	if err != nil {
		return err
	}
	apiAcls := make(map[int32]*lb.ACL)
	for _, acl := range resACL.ACLs {
		apiAcls[acl.Index] = acl
	}

	//convert state acl and sanitize them a bit
	newACL := make([]*lb.ACL, 0)
	for _, rawACL := range d.Get("acl").([]interface{}) {
		newACL = append(newACL, expandLbACL(rawACL))
	}

	//loop
	for index, stateACL := range newACL {
		key := int32(index) + 1
		if apiACL, found := apiAcls[key]; found {
			//there is an old acl with the same key. Remove it from array to mark that we've dealt with it
			delete(apiAcls, key)

			//if the state acl doesn't specify a name, set it to the same as the existing rule
			if stateACL.Name == "" {
				stateACL.Name = apiACL.Name
			}
			//Verify if their values are the same and ignore if that's the case, update otherwise
			if aclEquals(stateACL, apiACL) {
				continue
			}
			_, err = lbAPI.UpdateACL(&lb.UpdateACLRequest{
				Region: region,
				ACLID:  apiACL.ID,
				Name:   stateACL.Name,
				Action: stateACL.Action,
				Match:  stateACL.Match,
				Index:  key,
			})
			if err != nil {
				return err
			}
			continue
		}
		//old acl doesn't exist, create a new one
		_, err = lbAPI.CreateACL(&lb.CreateACLRequest{
			Region:     region,
			FrontendID: frontendID,
			Name:       expandOrGenerateString(stateACL.Name, "lb-acl"),
			Action:     stateACL.Action,
			Match:      stateACL.Match,
			Index:      key,
		})
		if err != nil {
			return err
		}
	}
	//we've finished with all new acl, delete any remaining old one which were not dealt with yet
	for _, acl := range apiAcls {
		err = lbAPI.DeleteACL(&lb.DeleteACLRequest{
			Region: region,
			ACLID:  acl.ID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceScalewayLbFrontendBetaUpdate(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := lbAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return err
	}

	req := &lb.UpdateFrontendRequest{
		Region:        region,
		FrontendID:    ID,
		Name:          d.Get("name").(string),
		InboundPort:   int32(d.Get("inbound_port").(int)),
		BackendID:     expandID(d.Get("backend_id")),
		TimeoutClient: expandDuration(d.Get("timeout_client")),
		CertificateID: expandStringPtr(expandID(d.Get("certificate_id"))),
	}

	_, err = lbAPI.UpdateFrontend(req)
	if err != nil {
		return err
	}

	//update acl
	err = resourceScalewayLbFrontendBetaUpdateACL(d, lbAPI, region, ID)
	if err != nil {
		return err
	}

	return resourceScalewayLbFrontendBetaRead(d, m)
}

func resourceScalewayLbFrontendBetaDelete(d *schema.ResourceData, m interface{}) error {
	lbAPI, region, ID, err := lbAPIWithRegionAndID(m, d.Id())
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

	return nil
}

func aclEquals(aclA, aclB *lb.ACL) bool {
	if aclA.Name != aclB.Name {
		return false
	}
	if !cmp.Equal(aclA.Match, aclB.Match) {
		return false
	}
	if !cmp.Equal(aclA.Action, aclB.Action) {
		return false
	}
	return true
}
