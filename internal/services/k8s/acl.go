package k8s

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceACL() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceACLCreate,
		ReadContext:   ResourceACLRead,
		UpdateContext: ResourceACLUpdate,
		DeleteContext: ResourceACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create:  schema.DefaultTimeout(defaultK8SClusterTimeout),
			Read:    schema.DefaultTimeout(defaultK8SClusterTimeout),
			Update:  schema.DefaultTimeout(defaultK8SClusterTimeout),
			Delete:  schema.DefaultTimeout(defaultK8SClusterTimeout),
			Default: schema.DefaultTimeout(defaultK8SClusterTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
				Description:      "Cluster on which the ACL is applied",
			},
			"acls": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The list of network rules that manage inbound traffic",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The IP subnet to be allowed",
						},
						"scaleway_ranges": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Allow access to cluster from all Scaleway ranges as defined in https://www.scaleway.com/en/docs/console/account/reference-content/scaleway-network-information/#ip-ranges-used-by-scaleway. Only one rule with this field set to true can be added",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The description of the ACL rule",
						},
					},
				},
			},
			// Common
			"region": regional.Schema(),
		},
		CustomizeDiff: cdf.LocalityCheck("cluster_id"),
	}
}

func ResourceACLCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	clusterID := d.Get("cluster_id").(string)

	_, err = waitCluster(ctx, api, region, locality.ExpandID(clusterID), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	acls, err := expandACL(d.Get("acls").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}

	createReq := &k8s.SetClusterACLRulesRequest{
		Region:    region,
		ClusterID: clusterID,
		ACLs:      acls,
	}

	_, err = api.SetClusterACLRules(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(clusterID)

	return ResourceACLRead(ctx, d, m)
}

func ResourceACLRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, clusterID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitCluster(ctx, api, region, clusterID, d.Timeout(schema.TimeoutRead))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	acls, err := api.ListClusterACLRules(&k8s.ListClusterACLRulesRequest{
		Region:    region,
		ClusterID: clusterID,
	}, scw.WithContext(ctx))
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	id := regional.NewID(region, clusterID).String()
	d.SetId(id)

	_ = d.Set("cluster_id", clusterID)
	_ = d.Set("acls", flattenACL(acls.Rules))

	return nil
}

func ResourceACLUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, clusterID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitCluster(ctx, api, region, clusterID, d.Timeout(schema.TimeoutUpdate))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	if d.HasChange("acls") {
		acls, err := expandACL(d.Get("acls").([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}

		req := &k8s.SetClusterACLRulesRequest{
			Region:    region,
			ClusterID: clusterID,
			ACLs:      acls,
		}
		_, err = api.SetClusterACLRules(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return ResourceACLRead(ctx, d, m)
}

func ResourceACLDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api, region, clusterID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitCluster(ctx, api, region, clusterID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	req := &k8s.SetClusterACLRulesRequest{
		Region:    region,
		ClusterID: clusterID,
		ACLs:      nil,
	}
	_, err = api.SetClusterACLRules(req, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitCluster(ctx, api, region, clusterID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	return nil
}

func expandACL(data []interface{}) ([]*k8s.ACLRuleRequest, error) {
	expandedACLs := []*k8s.ACLRuleRequest(nil)

	for _, rule := range data {
		r := rule.(map[string]interface{})
		expandedRule := &k8s.ACLRuleRequest{}

		if ipRaw, ok := r["ip"]; ok {
			ip, err := types.ExpandIPNet(ipRaw.(string))
			if err != nil {
				return nil, err
			}
			expandedRule.IP = &ip
		}
		if scwRangesRaw, ok := r["scaleway_ranges"]; ok {
			expandedRule.ScalewayRanges = scw.BoolPtr(scwRangesRaw.(bool))
		}
		if descriptionRaw, ok := r["description"]; ok {
			expandedRule.Description = descriptionRaw.(string)
		}

		expandedACLs = append(expandedACLs, expandedRule)
	}

	return expandedACLs, nil
}

func flattenACL(rules []*k8s.ACLRule) interface{} {
	if rules == nil {
		return nil
	}

	flattenedACLs := []map[string]interface{}(nil)
	for _, rule := range rules {
		flattenedACLs = append(flattenedACLs, map[string]interface{}{
			//"id": rule.ID,
			"ip":              rule.IP,
			"scaleway_ranges": rule.ScalewayRanges,
			"description":     rule.Description,
		})
	}

	return flattenedACLs
}
