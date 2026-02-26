package instance

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/httperrors"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/identity"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/locality/zonal"
	"github.com/scaleway/terraform-provider-scaleway/v2/internal/verify"
)

func ResourceSecurityGroupRules() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceInstanceSecurityGroupRulesCreate,
		ReadContext:   ResourceInstanceSecurityGroupRulesRead,
		UpdateContext: ResourceInstanceSecurityGroupRulesUpdate,
		DeleteContext: ResourceInstanceSecurityGroupRulesDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Default: schema.DefaultTimeout(defaultInstanceSecurityGroupRuleTimeout),
		},
		SchemaFunc: securityGroupRulesSchema,
		Identity:   identity.DefaultZonal(),
	}
}

func securityGroupRulesSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"security_group_id": {
			Type:     schema.TypeString,
			Required: true,
			// Ensure SecurityGroupRules.ID and SecurityGroupRules.security_group_id stay in sync.
			// If security_group_id is changed, a new SecurityGroupRules is created, with a new ID.
			ForceNew:         true,
			Description:      "The security group associated with this volume",
			ValidateDiagFunc: verify.IsUUIDWithLocality(),
		},
		"inbound_rule": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Inbound rules for this set of security group rules",
			Elem:        securityGroupRuleSchema(),
		},
		"outbound_rule": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Outbound rules for this set of security group rules",
			Elem:        securityGroupRuleSchema(),
		},
	}
}

func ResourceInstanceSecurityGroupRulesCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	zone, securityGroupID, err := locality.ParseLocalizedID(d.Get("security_group_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetZonalIdentity(d, scw.Zone(zone), securityGroupID)
	if err != nil {
		return diag.FromErr(err)
	}

	// We call update instead of read as it will take care of creating rules.
	return ResourceInstanceSecurityGroupRulesUpdate(ctx, d, m)
}

func setSecurityGroupRulesState(ctx context.Context, d *schema.ResourceData, instanceAPI *instance.API, sg *instance.SecurityGroup) diag.Diagnostics {
	_ = d.Set("security_group_id", zonal.NewID(sg.Zone, sg.ID).String())

	inboundRules, outboundRules, err := getSecurityGroupRules(ctx, instanceAPI, sg.Zone, sg.ID, d)
	if err != nil {
		if httperrors.Is404(err) {
			d.SetId("")

			return nil
		}

		return diag.FromErr(err)
	}

	_ = d.Set("inbound_rule", inboundRules)
	_ = d.Set("outbound_rule", outboundRules)

	return nil
}

func ResourceInstanceSecurityGroupRulesRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	securityGroupZonedID := d.Id()

	instanceAPI, zone, securityGroupID, err := NewAPIWithZoneAndID(m, securityGroupZonedID)
	if err != nil {
		return diag.FromErr(err)
	}

	sg, err := instanceAPI.GetSecurityGroup(&instance.GetSecurityGroupRequest{
		Zone:            zone,
		SecurityGroupID: securityGroupID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	err = identity.SetZonalIdentity(d, sg.SecurityGroup.Zone, sg.SecurityGroup.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	return setSecurityGroupRulesState(ctx, d, instanceAPI, sg.SecurityGroup)
}

func ResourceInstanceSecurityGroupRulesUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	securityGroupZonedID := d.Id()

	instanceAPI, zone, securityGroupID, err := NewAPIWithZoneAndID(m, securityGroupZonedID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = updateSecurityGroupeRules(ctx, d, zone, securityGroupID, instanceAPI)
	if err != nil {
		return diag.FromErr(err)
	}

	sg, err := instanceAPI.GetSecurityGroup(&instance.GetSecurityGroupRequest{
		Zone:            zone,
		SecurityGroupID: securityGroupID,
	}, scw.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	return setSecurityGroupRulesState(ctx, d, instanceAPI, sg.SecurityGroup)
}

func ResourceInstanceSecurityGroupRulesDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	securityGroupZonedID := d.Id()

	instanceAPI, zone, securityGroupID, err := NewAPIWithZoneAndID(m, securityGroupZonedID)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("inbound_rule", nil)
	_ = d.Set("outbound_rule", nil)

	err = updateSecurityGroupeRules(ctx, d, zone, securityGroupID, instanceAPI)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
