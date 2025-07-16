package lb

import (
	"context"
	"math"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceFrontend() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLbFrontendCreate,
		ReadContext:   resourceLbFrontendRead,
		UpdateContext: resourceLbFrontendUpdate,
		DeleteContext: resourceLbFrontendDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultLbLbTimeout),
			Update:  schema.DefaultTimeout(defaultLbLbTimeout),
			Delete:  schema.DefaultTimeout(defaultLbLbTimeout),
			Default: schema.DefaultTimeout(defaultLbLbTimeout),
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{Version: 0, Type: lbUpgradeV1SchemaType(), Upgrade: UpgradeStateV1Func},
		},
		Schema: map[string]*schema.Schema{
			"lb_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				Description:      "The load-balancer ID",
			},
			"backend_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				Description:      "The load-balancer backend ID",
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
				DiffSuppressFunc: dsf.Duration,
				ValidateDiagFunc: verify.IsDuration(),
				Description:      "Set the maximum inactivity time on the client side",
			},
			"certificate_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Certificate ID",
				Deprecated:  "Please use certificate_ids",
			},
			"certificate_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				},
				Description:      "Collection of Certificate IDs related to the load balancer and domain",
				DiffSuppressFunc: dsf.OrderDiff,
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
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description of the ACL",
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
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: verify.ValidateEnum[lbSDK.ACLActionType](),
										Description:      "The action type",
									},
									"redirect": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "Redirect parameters when using an ACL with `redirect` action",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: verify.ValidateEnum[lbSDK.ACLActionRedirectRedirectType](),
													Description:      "The redirect type",
												},
												"target": {
													Type:        schema.TypeString,
													Optional:    true,
													Description: "An URL can be used in case of a location redirect ",
												},
												"code": {
													Type:        schema.TypeInt,
													Optional:    true,
													Description: "The HTTP redirect code to use",
												},
											},
										},
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
										Optional:         true,
										Description:      "A list of IPs or CIDR v4/v6 addresses of the client of the session to match",
										DiffSuppressFunc: diffSuppressFunc32SubnetMask,
									},
									"http_filter": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          lbSDK.ACLHTTPFilterACLHTTPFilterNone.String(),
										ValidateDiagFunc: verify.ValidateEnum[lbSDK.ACLHTTPFilter](),
										Description:      "The HTTP filter to match",
									},
									"http_filter_value": {
										Type:        schema.TypeList,
										Optional:    true,
										Description: "A list of possible values to match for the given HTTP filter",
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"http_filter_option": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "You can use this field with http_header_match acl type to set the header name to filter",
									},
									"invert": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: `If set to true, the condition will be of type "unless"`,
									},
									"ips_edge_services": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: `Defines whether Edge Services IPs should be matched`,
									},
								},
							},
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IsDate and time of ACL's creation (RFC 3339 format)",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IsDate and time of ACL's update (RFC 3339 format)",
						},
					},
				},
			},
			"external_acls": {
				Type:          schema.TypeBool,
				Description:   "This boolean determines if ACLs should be managed externally through the 'lb_acl' resource. If set to `true`, `acl` attribute cannot be set directly in the lb frontend",
				Optional:      true,
				Default:       false,
				ConflictsWith: []string{"acl"},
			},
			"enable_http3": {
				Type:        schema.TypeBool,
				Description: "Activates HTTP/3 protocol",
				Optional:    true,
				Default:     false,
			},
			"connection_rate_limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Rate limit for new connections established on this frontend. Use 0 value to disable, else value is connections per second",
			},
			"enable_access_logs": {
				Type:        schema.TypeBool,
				Description: "Defines whether to enable access logs on the frontend",
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func resourceLbFrontendCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lbAPI, _, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	lbID := locality.ExpandID(d.Get("lb_id"))
	if lbID == "" {
		return diag.Errorf("load balancer id wrong format: %v", d.Get("lb_id").(string))
	}

	// parse lb_id. It will be forced to a zoned lb
	zone, _, err := zonal.ParseID(d.Get("lb_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	backZone, _, err := zonal.ParseID(d.Get("backend_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	if zone != backZone {
		return diag.Errorf("Frontend and Backend must be in the same zone (got %s and %s)", zone, backZone)
	}

	_, err = waitForLB(ctx, lbAPI, zone, lbID, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		if httperrors.Is403(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	timeoutClient, err := types.ExpandDuration(d.Get("timeout_client"))
	if err != nil {
		return diag.FromErr(err)
	}

	createFrontendRequest := &lbSDK.ZonedAPICreateFrontendRequest{
		Zone:                zone,
		LBID:                lbID,
		Name:                types.ExpandOrGenerateString(d.Get("name"), "lb-frt"),
		InboundPort:         int32(d.Get("inbound_port").(int)),
		BackendID:           locality.ExpandID(d.Get("backend_id")),
		TimeoutClient:       timeoutClient,
		EnableHTTP3:         d.Get("enable_http3").(bool),
		ConnectionRateLimit: types.ExpandUint32Ptr(d.Get("connection_rate_limit")),
		EnableAccessLogs:    d.Get("enable_access_logs").(bool),
	}

	certificatesRaw, certificatesExist := d.GetOk("certificate_ids")
	if certificatesExist {
		createFrontendRequest.CertificateIDs = types.ExpandSliceIDsPtr(certificatesRaw)
	}

	frontend, err := lbAPI.CreateFrontend(createFrontendRequest, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(zone, frontend.ID))

	if d.Get("external_acls").(bool) {
		return resourceLbFrontendRead(ctx, d, m)
	}

	return resourceLbFrontendUpdate(ctx, d, m)
}

func resourceLbFrontendRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lbAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	frontend, err := lbAPI.GetFrontend(&lbSDK.ZonedAPIGetFrontendRequest{
		Zone:       zone,
		FrontendID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("lb_id", zonal.NewIDString(zone, frontend.LB.ID))
	_ = d.Set("backend_id", zonal.NewIDString(zone, frontend.Backend.ID))
	_ = d.Set("name", frontend.Name)
	_ = d.Set("inbound_port", int(frontend.InboundPort))
	_ = d.Set("timeout_client", types.FlattenDuration(frontend.TimeoutClient))
	_ = d.Set("enable_http3", frontend.EnableHTTP3)
	_ = d.Set("connection_rate_limit", types.FlattenUint32Ptr(frontend.ConnectionRateLimit))
	_ = d.Set("enable_access_logs", frontend.EnableAccessLogs)

	if frontend.Certificate != nil { //nolint:staticcheck
		_ = d.Set("certificate_id", zonal.NewIDString(zone, frontend.Certificate.ID)) //nolint:staticcheck
	} else {
		_ = d.Set("certificate_id", "")
	}

	if len(frontend.CertificateIDs) > 0 {
		_ = d.Set("certificate_ids", types.FlattenSliceIDs(frontend.CertificateIDs, zone))
	}

	if !d.Get("external_acls").(bool) {
		// read related acls.
		resACL, err := lbAPI.ListACLs(&lbSDK.ZonedAPIListACLsRequest{
			Zone:       zone,
			FrontendID: ID,
		}, scw.WithAllPages(), scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}

		_ = d.Set("acl", flattenLBACLs(resACL.ACLs))
	}

	return nil
}

func flattenLBACLs(acls []*lbSDK.ACL) any {
	sort.Slice(acls, func(i, j int) bool {
		return acls[i].Index < acls[j].Index
	})

	rawACLs := make([]any, 0, len(acls))
	for _, apiACL := range acls {
		rawACLs = append(rawACLs, flattenLbACL(apiACL))
	}

	return rawACLs
}

func resourceLbFrontendUpdateACL(ctx context.Context, d *schema.ResourceData, lbAPI *lbSDK.ZonedAPI, zone scw.Zone, frontendID string) diag.Diagnostics {
	// Fetch existing acl from the api. and convert it to a hashmap with index as key
	resACL, err := lbAPI.ListACLs(&lbSDK.ZonedAPIListACLsRequest{
		Zone:       zone,
		FrontendID: frontendID,
	}, scw.WithAllPages(), scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	apiACLs := make(map[int32]*lbSDK.ACL)
	for _, acl := range resACL.ACLs {
		apiACLs[acl.Index] = acl
	}

	// convert state acl and sanitize them a bit
	newACL := expandsLBACLs(d, d.Get("acl"))

	// loop
	for index, stateACL := range newACL {
		key := int32(index) + 1
		if apiACL, found := apiACLs[key]; found {
			// there is an old acl with the same key. Remove it from array to mark that we've dealt with it
			delete(apiACLs, key)

			// if the state acl doesn't specify a name, set it to the same as the existing rule
			if stateACL.Name == "" {
				stateACL.Name = apiACL.Name
			}
			// Verify if their values are the same and ignore if that's the case, update otherwise
			if ACLEquals(stateACL, apiACL) {
				continue
			}

			_, err = lbAPI.UpdateACL(&lbSDK.ZonedAPIUpdateACLRequest{
				Zone:   zone,
				ACLID:  apiACL.ID,
				Name:   stateACL.Name,
				Action: stateACL.Action,
				Match:  stateACL.Match,
				Index:  key,
			})
			if err != nil {
				return diag.FromErr(err)
			}

			continue
		}
		// old acl doesn't exist, create a new one
		_, err = lbAPI.CreateACL(&lbSDK.ZonedAPICreateACLRequest{
			Zone:       zone,
			FrontendID: frontendID,
			Name:       types.ExpandOrGenerateString(stateACL.Name, "lb-acl"),
			Action:     stateACL.Action,
			Match:      stateACL.Match,
			Index:      key,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	// we've finished with all new acl, delete any remaining old one which were not dealt with yet
	for _, acl := range apiACLs {
		err = lbAPI.DeleteACL(&lbSDK.ZonedAPIDeleteACLRequest{
			Zone:  zone,
			ACLID: acl.ID,
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func expandsLBACLs(d *schema.ResourceData, raw any) []*lbSDK.ACL {
	r := raw.([]any)
	newACL := make([]*lbSDK.ACL, 0)

	for index, rawACL := range r {
		newACL = append(newACL, expandLbACL(d, rawACL, index))
	}

	return newACL
}

func resourceLbFrontendUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lbAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, lbID, err := zonal.ParseID(d.Get("lb_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// check err waiting process
	_, err = waitForLB(ctx, lbAPI, zone, lbID, d.Timeout(schema.TimeoutUpdate))
	if err != nil {
		if httperrors.Is403(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	timeoutClient, err := types.ExpandDuration(d.Get("timeout_client"))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &lbSDK.ZonedAPIUpdateFrontendRequest{
		Zone:                zone,
		FrontendID:          ID,
		Name:                types.ExpandOrGenerateString(d.Get("name"), "lb-frt"),
		InboundPort:         int32(d.Get("inbound_port").(int)),
		BackendID:           locality.ExpandID(d.Get("backend_id")),
		TimeoutClient:       timeoutClient,
		CertificateIDs:      types.ExpandSliceIDsPtr(d.Get("certificate_ids")),
		EnableHTTP3:         d.Get("enable_http3").(bool),
		ConnectionRateLimit: types.ExpandUint32Ptr(d.Get("connection_rate_limit")),
		EnableAccessLogs:    types.ExpandBoolPtr(d.Get("enable_access_logs")),
	}

	_, err = lbAPI.UpdateFrontend(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	diagnostics := resourceLbFrontendUpdateACL(ctx, d, lbAPI, zone, ID)
	if diagnostics != nil {
		return diagnostics
	}

	return resourceLbFrontendRead(ctx, d, m)
}

func resourceLbFrontendDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lbAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, lbID, err := zonal.ParseID(d.Get("lb_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	err = lbAPI.DeleteFrontend(&lbSDK.ZonedAPIDeleteFrontendRequest{
		Zone:       zone,
		FrontendID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForLB(ctx, lbAPI, zone, lbID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}

func ACLEquals(aclA, aclB *lbSDK.ACL) bool {
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
