package k8s

import (
	"context"
	_ "embed"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/k8s/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/cdf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/dsf"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/regional"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/types"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

//go:embed descriptions/acl.md
var aclDescription string

func ResourceACL() *schema.Resource {
	return &schema.Resource{
		Description:   aclDescription,
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
		SchemaFunc:    aclSchema,
		CustomizeDiff: cdf.LocalityCheck("cluster_id"),
	}
}

func aclSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cluster_id": {
			Type:             schema.TypeString,
			Required:         true,
			ForceNew:         true,
			ValidateDiagFunc: verify.IsUUIDorUUIDWithLocality(),
			DiffSuppressFunc: dsf.Locality,
			Description:      "Cluster on which the ACL should be applied",
		},
		"no_ip_allowed": {
			Type:         schema.TypeBool,
			Optional:     true,
			Default:      false,
			Description:  "If true, no IP will be allowed and the cluster will be fully isolated",
			ExactlyOneOf: []string{"acl_rules"},
		},
		"acl_rules": {
			Type:         schema.TypeSet,
			Optional:     true,
			Description:  "The list of network rules that manage inbound traffic",
			ExactlyOneOf: []string{"no_ip_allowed"},
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"ip": {
						Type:         schema.TypeString,
						Optional:     true,
						Description:  "The IP subnet to be allowed",
						ValidateFunc: validation.IsCIDR,
					},
					"scaleway_ranges": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "Allow access to cluster from all Scaleway ranges",
					},
					"description": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The description of the ACL rule",
					},
					"id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The ID of the ACL rule",
					},
				},
			},
		},
		// Common
		"region": regional.Schema(),
	}
}

func ResourceACLCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, _, err := newAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	region, clusterID, err := regional.ParseID(d.Get("cluster_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitCluster(ctx, api, region, locality.ExpandID(clusterID), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	acls, err := expandACL(d.Get("acl_rules"))
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

	regionalID := regional.NewID(region, clusterID).String()
	d.SetId(regionalID)

	return ResourceACLRead(ctx, d, m)
}

func ResourceACLRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	_ = d.Set("cluster_id", regional.NewIDString(region, clusterID))
	_ = d.Set("region", region)
	_ = d.Set("acl_rules", flattenACL(acls.Rules))

	return nil
}

func ResourceACLUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, clusterID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitCluster(ctx, api, region, clusterID, d.Timeout(schema.TimeoutUpdate))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	if d.HasChanges("acl_rules", "no_ip_allowed") {
		acls, err := expandACL(d.Get("acl_rules"))
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

func ResourceACLDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	api, region, clusterID, err := NewAPIWithRegionAndID(m, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitCluster(ctx, api, region, clusterID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !httperrors.Is404(err) {
		return diag.FromErr(err)
	}

	allowedIPs, err := types.ExpandIPNet("0.0.0.0/0")
	if err != nil {
		return diag.FromErr(err)
	}

	req := &k8s.SetClusterACLRulesRequest{
		Region:    region,
		ClusterID: clusterID,
		ACLs: []*k8s.ACLRuleRequest{
			{
				IP:          &allowedIPs,
				Description: "Automatically generated after scaleway_k8s_acl resource deletion",
			},
		},
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

func expandACL(data any) ([]*k8s.ACLRuleRequest, error) {
	expandedACLs := []*k8s.ACLRuleRequest(nil)

	if data == nil {
		return expandedACLs, nil
	}

	for _, rule := range data.(*schema.Set).List() {
		r := rule.(map[string]any)
		expandedRule := &k8s.ACLRuleRequest{}

		if ipRaw, ok := r["ip"]; ok && ipRaw != "" {
			ip, err := types.ExpandIPNet(ipRaw.(string))
			if err != nil {
				return nil, err
			}

			expandedRule.IP = &ip
		}

		if scwRangesRaw, ok := r["scaleway_ranges"]; ok && scwRangesRaw.(bool) {
			expandedRule.ScalewayRanges = new(true)
		}

		if descriptionRaw, ok := r["description"]; ok && descriptionRaw.(string) != "" {
			expandedRule.Description = descriptionRaw.(string)
		}

		expandedACLs = append(expandedACLs, expandedRule)
	}

	return expandedACLs, nil
}

func flattenACL(rules []*k8s.ACLRule) any {
	if rules == nil {
		return nil
	}

	flattenedACLs := []map[string]any(nil)

	for _, rule := range rules {
		flattenedRule := map[string]any{
			"id":              rule.ID,
			"scaleway_ranges": rule.ScalewayRanges,
			"description":     rule.Description,
		}
		if rule.IP != nil {
			flattenedRule["ip"] = rule.IP.String()
		}

		flattenedACLs = append(flattenedACLs, flattenedRule)
	}

	return flattenedACLs
}
