package lb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
)

func DataSourceACLs() *schema.Resource {
	return &schema.Resource{
		ReadContext: DataSourceLbACLsRead,
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
				Type:        schema.TypeList,
				Computed:    true,
				Description: "ACLs that are listed.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Computed:    true,
							Description: "UUID of the ACL.",
							Type:        schema.TypeString,
						},
						"name": {
							Computed:    true,
							Description: "Name of the ACL.",
							Type:        schema.TypeString,
						},
						"frontend_id": {
							Computed:    true,
							Description: "ID of the frontend to use for the ACL.",
							Type:        schema.TypeString,
						},
						"index": {
							Computed:    true,
							Description: "Priority of this ACL (ACLs are applied in ascending order, 0 is the first ACL executed).",
							Type:        schema.TypeInt,
						},
						"description": {
							Computed:    true,
							Description: "ACL description.",
							Type:        schema.TypeString,
						},
						"match": {
							Type:        schema.TypeList,
							Description: "ACL Match configuration.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"ip_subnet": {
										Computed:    true,
										Description: "List of IPs or CIDR v4/v6 addresses to filter for from the client side.",
										Type:        schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"http_filter": {
										Type:        schema.TypeString,
										Description: "type of HTTP filter to match. Extracts the request's URL path, which starts at the first slash and ends before the question mark (without the host part). Defines where to filter for the http_filter_value. Only supported for HTTP backends.",
										Computed:    true,
									},
									"http_filter_value": {
										Computed:    true,
										Description: "List of values to filter for",
										Type:        schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"http_filter_option": {
										Type:        schema.TypeString,
										Description: "Name of the HTTP header to filter on if `http_header_match` was selected in `http_filter`.",
										Computed:    true,
									},
									"invert": {
										Computed:    true,
										Description: "Defines whether to invert the match condition. If set to `true`, the ACL carries out its action when the condition DOES NOT match",
										Type:        schema.TypeBool,
									},
									"ips_edge_services": {
										Computed:    true,
										Description: "Defines whether Edge Services IPs should be matched",
										Type:        schema.TypeBool,
									},
								},
							},
						},
						"action": {
							Type:        schema.TypeList,
							Description: "Action to take when incoming traffic matches an ACL filter.",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:        schema.TypeString,
										Description: "type: action to take when incoming traffic matches an ACL filter. (allow/deny)",
										Computed:    true,
									},
									"redirect": {
										Type:        schema.TypeList,
										Description: "redirection parameters when using an ACL with a `redirect` action.",
										Computed:    true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:        schema.TypeString,
													Description: "Value can be location or scheme",
													Computed:    true,
												},
												"target": {
													Type:        schema.TypeString,
													Description: "redirect target. For a location redirect, you can use a URL e.g. `https://scaleway.com`. Using a scheme name (e.g. `https`, `http`, `ftp`, `git`) will replace the request's original scheme. This can be useful to implement HTTP to HTTPS redirects. Valid placeholders that can be used in a `location` redirect to preserve parts of the original request in the redirection URL are {{host}}, {{query}}, {{path}} and {{scheme}}.",
													Computed:    true,
												},
												"code": {
													Type:        schema.TypeInt,
													Description: "HTTP redirect code to use. Valid values are 301, 302, 303, 307 and 308. Default value is 302.",
													Computed:    true,
												},
											},
										},
									},
								},
							},
						},
						"created_at": {
							Computed:    true,
							Description: "Timestamp when the ACL was created (RFC3339)",
							Type:        schema.TypeString,
						},
						"update_at": {
							Computed:    true,
							Description: "Timestamp when the ACL was updated (RFC3339)",
							Type:        schema.TypeString,
						},
					},
				},
			},
			"zone":            zonal.Schema(),
			"organization_id": account.OrganizationIDSchema(),
			"project_id":      account.ProjectIDSchema(),
		},
	}
}

func DataSourceLbACLsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	lbAPI, zone, err := lbAPIWithZone(d, m)
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
		Name:       types.ExpandStringPtr(d.Get("name")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	acls := []any(nil)

	for _, acl := range res.ACLs {
		rawACL := make(map[string]any)
		rawACL["id"] = zonal.NewIDString(zone, acl.ID)
		rawACL["name"] = acl.Name
		rawACL["frontend_id"] = zonal.NewIDString(zone, acl.Frontend.ID)
		rawACL["created_at"] = types.FlattenTime(acl.CreatedAt)
		rawACL["update_at"] = types.FlattenTime(acl.UpdatedAt)
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
