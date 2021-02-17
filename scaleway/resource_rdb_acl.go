package scaleway

import (
	"context"
	"fmt"
	"strings"

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
							ValidateFunc: validation.IsIPAddress,
							Required:     true,
							Description:  "Target IP of the rules",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
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

func resourceScalewayRdbACLCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}
	instanceID := d.Get("instance_id").(string)
	createReq := &rdb.SetInstanceACLRulesRequest{
		Region:     region,
		InstanceID: instanceID,
		Rules:      rdbACLExpand(d.Get("acl_rules")),
	}

	_, err = rdbAPI.SetInstanceACLRules(createReq, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resourceScalewayRdbACLID(region, expandID(instanceID)))

	return resourceScalewayRdbACLRead(ctx, d, m)
}

func resourceScalewayRdbACLRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID, err := resourceScalewayRdbACLParseID(d.Id())

	if err != nil {
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

	_ = d.Set("instance_id", newRegionalID(region, instanceID).String())
	_ = d.Set("acl_rules", rdbACLRulesFlatten(res.Rules))

	return nil
}

func resourceScalewayRdbACLUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	instanceID, err := resourceScalewayRdbACLParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("ip_rules") {
		req := &rdb.SetInstanceACLRulesRequest{
			Region:     region,
			InstanceID: instanceID,
			Rules:      rdbACLExpand(d.Get("acl_rules")),
		}

		_, err = rdbAPI.SetInstanceACLRules(req, scw.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceScalewayRdbACLRead(ctx, d, m)
}

func resourceScalewayRdbACLDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	rdbAPI, region, err := rdbAPIWithRegion(d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = rdbAPI.DeleteInstanceACLRules(&rdb.DeleteInstanceACLRulesRequest{
		Region: region,
	}, scw.WithContext(ctx))

	if err != nil && !is404Error(err) {
		return diag.FromErr(err)
	}

	return nil
}

// Build the resource identifier
// The resource identifier format is "Region/InstanceId/UserName"
func resourceScalewayRdbACLID(region scw.Region, instanceID string) (resourceID string) {
	return fmt.Sprintf("%s/%s", region, instanceID)
}

// The resource identifier format is "Region/InstanceId/acl"
func resourceScalewayRdbACLParseID(resourceID string) (instanceID string, err error) {
	idParts := strings.Split(resourceID, "/")
	if len(idParts) != 3 {
		return "", fmt.Errorf("can't parse user resource id: %s", resourceID)
	}
	return idParts[1], nil
}

func rdbACLExpand(data interface{}) []*rdb.ACLRuleRequest {
	type aclRule struct {
		IP          string
		Description string
	}
	var res []*rdb.ACLRuleRequest
	for _, rule := range data.([]interface{}) {
		r := rule.(map[string]interface{})
		res = append(res, &rdb.ACLRuleRequest{
			IP:          expandIPNet(r["ip"].(string)),
			Description: r["description"].(string),
		})
	}

	return res
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
