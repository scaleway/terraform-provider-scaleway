package scaleway

import (
	"bytes"
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

func resourceScalewayRdbACL() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalewayRdbACLCreate,
		ReadContext:   resourceScalewayRdbACLRead,
		UpdateContext: resourceScalewayRdbACLUpdate,
		DeleteContext: resourceScalewayRdbACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultRdbInstanceTimeout),
		},
		SchemaVersion: 0,
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validationUUIDorUUIDWithLocality(),
				Description:  "Instance on which the ACL is applied",
			},
			"acl_rules": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of ACL rules to apply",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:         schema.TypeString,
							ValidateFunc: validation.IsCIDR,
							Required:     true,
							Description:  "Target IP of the rules",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Description of the rule",
						},
					},
				},
			},
			// Common
			"region": regionSchema(),
		},
	}
}

func resourceScalewayRdbACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	instanceID := d.Get("instance_id").(string)
	rdbAPI, region, ID, err := rdbAPIWithRegionAndID(meta, instanceID)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForRDBInstance(ctx, rdbAPI, region, expandID(instanceID), d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	aclRules, err := rdbACLExpand(d.Get("acl_rules").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}
	createReq := &rdb.SetInstanceACLRulesRequest{
		Region:     region,
		InstanceID: ID,
		Rules:      aclRules,
	}

	_, err = rdbAPI.SetInstanceACLRules(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(instanceID)

	return resourceScalewayRdbACLRead(ctx, d, meta)
}

func resourceScalewayRdbACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI, region, instanceID, err := rdbAPIWithRegionAndID(meta, d.Get("instance_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForRDBInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutRead))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	res, err := rdbAPI.ListInstanceACLRules(&rdb.ListInstanceACLRulesRequest{
		Region:     region,
		InstanceID: instanceID,
	}, scw.WithContext(ctx))

	if err != nil {
		if is404Error(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	id := newRegionalID(region, instanceID).String()
	d.SetId(id)
	_ = d.Set("instance_id", id)
	_ = d.Set("acl_rules", rdbACLRulesFlatten(res.Rules, d.Get("acl_rules").([]interface{})))

	return nil
}

func resourceScalewayRdbACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI, region, instanceID, err := rdbAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForRDBInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutUpdate))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	if d.HasChange("acl_rules") {
		_, err := waitForRDBInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}

		aclRules, err := rdbACLExpand(d.Get("acl_rules").([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
		req := &rdb.SetInstanceACLRulesRequest{
			Region:     region,
			InstanceID: instanceID,
			Rules:      aclRules,
		}

		_, err = rdbAPI.SetInstanceACLRules(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayRdbACLRead(ctx, d, meta)
}

func resourceScalewayRdbACLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI, region, instanceID, err := rdbAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	aclRuleIPs := make([]string, 0)
	aclRules, err := rdbACLExpand(d.Get("acl_rules").([]interface{}))
	if err != nil {
		return diag.FromErr(err)
	}
	for _, acl := range aclRules {
		aclRuleIPs = append(aclRuleIPs, acl.IP.String())
	}

	_, err = waitForRDBInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	_, err = rdbAPI.DeleteInstanceACLRules(&rdb.DeleteInstanceACLRulesRequest{
		Region:     region,
		InstanceID: instanceID,
		ACLRuleIPs: aclRuleIPs,
	}, scw.WithContext(ctx))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	_, err = waitForRDBInstance(ctx, rdbAPI, region, instanceID, d.Timeout(schema.TimeoutDelete))
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}

func rdbACLExpand(data []interface{}) ([]*rdb.ACLRuleRequest, error) {
	var res []*rdb.ACLRuleRequest
	for _, rule := range data {
		r := rule.(map[string]interface{})
		ip, err := expandIPNet(r["ip"].(string))
		if err != nil {
			return res, err
		}
		res = append(res, &rdb.ACLRuleRequest{
			IP:          ip,
			Description: r["description"].(string),
		})
	}

	sort.Slice(res, func(i, j int) bool {
		return bytes.Compare(res[i].IP.IP, res[j].IP.IP) < 0
	})

	return res, nil
}

func rdbACLRulesFlatten(rules []*rdb.ACLRule, data []interface{}) []map[string]interface{} {
	var res []map[string]interface{}
	m := make(map[string]*rdb.ACLRule)
	for _, rule := range rules {
		m[rule.IP.String()] = rule
	}

	for _, rule := range data {
		currentRule := rule.(map[string]interface{})
		ip, err := expandIPNet(currentRule["ip"].(string))
		if err != nil {
			return nil // append errors
		}

		aclRule := m[ip.String()]
		r := map[string]interface{}{
			"ip":          aclRule.IP.String(),
			"description": aclRule.Description,
		}
		// Add element on schema order
		res = append(res, r)
	}
	return res
}
