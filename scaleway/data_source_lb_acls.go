package scaleway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
)

func dataSourceScalewayLbACLs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceScalewayLbACLsRead,
		Schema: map[string]*schema.Schema{
			"frontend_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ACLs with a frontend id like it are listed.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ACLs with a name like it are listed.",
			},
			"acls": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"name": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"frontend_id": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"index": {
							Computed: true,
							Type:     schema.TypeInt,
						},
						"description": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"match": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip_subnet": {
										Computed: true,
										Type:     schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"http_filter": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"http_filter_value": {
										Computed: true,
										Type:     schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"http_filter_option": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"invert": {
										Computed: true,
										Type:     schema.TypeBool,
									},
								},
							},
						},
						"action": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"redirect": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"target": {
													Type:     schema.TypeString,
													Computed: true,
												},
												"code": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"created_at": {
							Computed: true,
							Type:     schema.TypeString,
						},
						"update_at": {
							Computed: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
			"zone":            zonal.Schema(),
			"organization_id": organizationIDSchema(),
			"project_id":      projectIDSchema(),
		},
	}
}

func dataSourceScalewayLbACLsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	_, frontID, err := zonal.ParseID(d.Get("frontend_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	res, err := lbAPI.ListACLs(&lb.ZonedAPIListACLsRequest{
		Zone:       zone,
		FrontendID: frontID,
		Name:       expandStringPtr(d.Get("name")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	acls := []interface{}(nil)
	for _, acl := range res.ACLs {
		rawACL := make(map[string]interface{})
		rawACL["id"] = zonal.NewIDString(zone, acl.ID)
		rawACL["name"] = acl.Name
		rawACL["frontend_id"] = zonal.NewIDString(zone, acl.Frontend.ID)
		rawACL["created_at"] = flattenTime(acl.CreatedAt)
		rawACL["update_at"] = flattenTime(acl.UpdatedAt)
		rawACL["index"] = acl.Index
		rawACL["description"] = acl.Description
		rawACL["action"] = flattenLbACLAction(acl.Action)
		rawACL["match"] = flattenLbACLMatch(acl.Match)

		acls = append(acls, rawACL)
	}

	d.SetId(zone.String())
	_ = d.Set("acls", acls)

	return nil
}
