package edgeservices

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	edgeservices "github.com/scaleway/scaleway-sdk-go/api/edge_services/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/services/account"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceRouteStage() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceRouteStageCreate,
		ReadContext:   ResourceRouteStageRead,
		UpdateContext: ResourceRouteStageUpdate,
		DeleteContext: ResourceRouteStageDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"pipeline_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the pipeline",
			},
			"waf_stage_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the WAF stage HTTP requests should be forwarded to when no rules are matched",
			},
			"rule": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of rules to be checked against every HTTP request. The first matching rule will forward the request to its specified backend stage. If no rules are matched, the request is forwarded to the WAF stage defined by `waf_stage_id`",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backend_stage_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the backend stage that requests matching the rule should be forwarded to",
						},
						"rule_http_match": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Rule condition to be matched. Requests matching the condition defined here will be directly forwarded to the backend specified by the `backend_stage_id` field. Requests that do not match will be checked by the next rule's condition",
							MaxItems:    1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"method_filters": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										Elem: &schema.Schema{
											Type:             schema.TypeString,
											ValidateDiagFunc: verify.ValidateEnum[edgeservices.RuleHTTPMatchMethodFilter](),
										},
										Description: "HTTP methods to filter for. A request using any of these methods will be considered to match the rule. Possible values are `get`, `post`, `put`, `patch`, `delete`, `head`, `options`. All methods will match if none is provided",
									},
									"path_filter": {
										Type:        schema.TypeList,
										Optional:    true,
										MaxItems:    1,
										Description: "HTTP URL path to filter for. A request whose path matches the given filter will be considered to match the rule. All paths will match if none is provided",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"path_filter_type": {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: verify.ValidateEnum[edgeservices.RuleHTTPMatchPathFilterPathFilterType](),
													Description:      "The type of filter to match for the HTTP URL path. For now, all path filters must be written in regex and use the `regex` type",
												},
												"value": {
													Type:        schema.TypeString,
													Required:    true,
													Description: "The value to be matched for the HTTP URL path",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the creation of the route stage",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time of the last update of the route stage",
			},
			"project_id": account.ProjectIDSchema(),
		},
	}
}

func ResourceRouteStageCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	routeStage, err := api.CreateRouteStage(&edgeservices.CreateRouteStageRequest{
		PipelineID: d.Get("pipeline_id").(string),
		WafStageID: types.ExpandStringPtr(d.Get("waf_stage_id").(string)),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = api.SetRouteRules(&edgeservices.SetRouteRulesRequest{
		RouteStageID: routeStage.ID,
		RouteRules:   expandRouteRules(d.Get("rule")),
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(routeStage.ID)

	return ResourceRouteStageRead(ctx, d, m)
}

func ResourceRouteStageRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	routeStage, err := api.GetRouteStage(&edgeservices.GetRouteStageRequest{
		RouteStageID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("pipeline_id", routeStage.PipelineID)
	_ = d.Set("waf_stage_id", types.FlattenStringPtr(routeStage.WafStageID))
	_ = d.Set("created_at", types.FlattenTime(routeStage.CreatedAt))
	_ = d.Set("updated_at", types.FlattenTime(routeStage.UpdatedAt))

	routeRules, err := api.ListRouteRules(&edgeservices.ListRouteRulesRequest{
		RouteStageID: routeStage.ID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("rule", flattenRouteRules(routeRules.RouteRules))

	return nil
}

func ResourceRouteStageUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	hasChanged := false

	updateRequest := &edgeservices.UpdateRouteStageRequest{
		RouteStageID: d.Id(),
	}

	if d.HasChange("waf_stage_id") {
		updateRequest.WafStageID = types.ExpandStringPtr(d.Get("waf_stage_id").(string))
		hasChanged = true
	}

	if hasChanged {
		_, err := api.UpdateRouteStage(updateRequest, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("rule") {
		_, err := api.SetRouteRules(&edgeservices.SetRouteRulesRequest{
			RouteStageID: d.Id(),
			RouteRules:   expandRouteRules(d.Get("rule")),
		}, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceRouteStageRead(ctx, d, m)
}

func ResourceRouteStageDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := NewEdgeServicesAPI(m)

	err := api.DeleteRouteStage(&edgeservices.DeleteRouteStageRequest{
		RouteStageID: d.Id(),
	}, scw.WithContext(ctx))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}
