package scaleway

import (
	"context"

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
				Type:        schema.TypeSet,
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

	_, err = waitInstance(ctx, rdbAPI, region, expandID(instanceID))
	if err != nil {
		return diag.FromErr(err)
	}

	aclRules, err := rdbACLExpand(d.Get("acl_rules").(*schema.Set))
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

	_, err = waitInstance(ctx, rdbAPI, region, instanceID)
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
	_ = d.Set("acl_rules", rdbACLRulesFlatten(res.Rules))

	return nil
}

func resourceScalewayRdbACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdbAPI, region, instanceID, err := rdbAPIWithRegionAndID(meta, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitInstance(ctx, rdbAPI, region, instanceID)
	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	if d.HasChange("acl_rules") {
		_ = rdb.WaitForInstanceRequest{
			InstanceID:    instanceID,
			Region:        region,
			RetryInterval: DefaultWaitRetryInterval,
		}

		aclRules, err := rdbACLExpand(d.Get("acl_rules").(*schema.Set))
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
	aclRules, err := rdbACLExpand(d.Get("acl_rules").(*schema.Set))
	if err != nil {
		return diag.FromErr(err)
	}
	for _, acl := range aclRules {
		aclRuleIPs = append(aclRuleIPs, acl.IP.String())
	}

	_, err = waitInstance(ctx, rdbAPI, region, instanceID)
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

	return nil
}

func rdbACLExpand(data *schema.Set) ([]*rdb.ACLRuleRequest, error) {
	var res []*rdb.ACLRuleRequest
	for _, rule := range data.List() {
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

	return res, nil
}

func rdbACLRulesFlatten(rules []*rdb.ACLRule) []map[string]interface{} {
	var res []map[string]interface{}
	for _, rule := range rules {
		r := map[string]interface{}{
			"ip":          rule.IP.String(),
			"description": rule.Description,
		}
		res = append(res, r)
	}
	return res
}
