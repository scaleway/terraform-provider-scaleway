package lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	lbSDK "github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceACL() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceLbACLCreate,
		ReadContext:   resourceLbACLRead,
		UpdateContext: resourceLbACLUpdate,
		DeleteContext: resourceLbACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultLbLbTimeout),
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"frontend_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				Description:      "The frontend ID on which the ACL is applied",
			},
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
			"index": {
				Type:         schema.TypeInt,
				Description:  "The priority of the ACL. (ACLs are applied in ascending order, 0 is the first ACL executed)",
				Required:     true,
				ValidateFunc: validation.IntAtLeast(0),
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
				Optional:    true,
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
	}
}

func resourceLbACLCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lbAPI, _, err := lbAPIWithZone(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	frontZone, frontID, err := zonal.ParseID(d.Get("frontend_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	req := &lbSDK.ZonedAPICreateACLRequest{
		Zone:        frontZone,
		FrontendID:  frontID,
		Name:        d.Get("name").(string),
		Action:      expandLbACLAction(d.Get("action")),
		Match:       expandLbACLMatch(d.Get("match")),
		Index:       int32(d.Get("index").(int)),
		Description: d.Get("description").(string),
	}

	res, err := lbAPI.CreateACL(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(zonal.NewIDString(frontZone, res.ID))

	return resourceLbACLRead(ctx, d, m)
}

func resourceLbACLRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lbAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	acl, err := lbAPI.GetACL(&lbSDK.ZonedAPIGetACLRequest{
		Zone:  zone,
		ACLID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("frontend_id", zonal.NewIDString(zone, acl.Frontend.ID))
	_ = d.Set("name", acl.Name)
	_ = d.Set("description", acl.Description)
	_ = d.Set("index", int(acl.Index))
	_ = d.Set("created_at", types.FlattenTime(acl.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(acl.UpdatedAt))
	_ = d.Set("action", flattenLbACLAction(acl.Action))

	if acl.Match != nil {
		_ = d.Set("match", flattenLbACLMatch(acl.Match))
	}

	return nil
}

func resourceLbACLUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lbAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	req := &lbSDK.ZonedAPIUpdateACLRequest{
		Zone:        zone,
		ACLID:       ID,
		Name:        d.Get("name").(string),
		Action:      expandLbACLAction(d.Get("action")),
		Index:       int32(d.Get("index").(int)),
		Match:       expandLbACLMatch(d.Get("match")),
		Description: types.ExpandUpdatedStringPtr(d.Get("description")),
	}

	_, err = lbAPI.UpdateACL(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceLbACLRead(ctx, d, m)
}

func resourceLbACLDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lbAPI, zone, ID, err := NewAPIWithZoneAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = lbAPI.DeleteACL(&lbSDK.ZonedAPIDeleteACLRequest{
		Zone:  zone,
		ACLID: ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
